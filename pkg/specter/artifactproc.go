package specter

import (
	"context"
	"encoding/json"
	"github.com/morebec/go-errors/errors"
	"io/fs"
	"os"
	"slices"
	"sync"
	"time"
)

// ArtifactRegistry provides an interface for managing a registry of artifacts. This
// registry tracks artifacts generated during processing runs, enabling clean-up
// in subsequent runs to avoid residual artifacts and maintain a clean slate.
//
// Implementations of the ArtifactRegistry interface must be thread-safe to handle
// concurrent calls to TrackFile and UntrackFile methods. Multiple goroutines may
// access the registry simultaneously, so appropriate synchronization mechanisms
// should be implemented to prevent race conditions and ensure data integrity.
type ArtifactRegistry interface {
	// Load the registry state from persistent storage. If an error occurs, it
	// should be returned to indicate the failure of the loading operation.
	Load() error

	// Save the current state of the registry to persistent storage. If an
	// error occurs, it should be returned to indicate the failure of the saving operation.
	Save() error

	// Add registers an ArtifactID under a specific processor name. This method
	// should ensure that the file path is associated with the given processor name
	// in the registry.
	Add(processorName string, e ArtifactRegistryEntry) error

	// Remove a given ArtifactID artifact registration for a specific processor name. This
	// method should ensure that the file path is disassociated from the given
	// processor name in the registry.
	Remove(processorName string, artifactID ArtifactID) error

	// FindByID finds an entry by its ArtifactID.
	FindByID(processorName string, artifactID ArtifactID) (entry ArtifactRegistryEntry, found bool, err error)

	// FindAll returns all the entries in the registry.
	FindAll(processorName string) ([]ArtifactRegistryEntry, error)
}

type ArtifactRegistryEntry struct {
	ArtifactID ArtifactID
	Metadata   map[string]any
}

// ProcessorArtifactRegistry is a wrapper around an ArtifactRegistry that scopes all calls to a given processor.
type ProcessorArtifactRegistry struct {
	processorName string
	registry      ArtifactRegistry
}

func NewProcessorArtifactRegistry(processorName string, registry ArtifactRegistry) ProcessorArtifactRegistry {
	return ProcessorArtifactRegistry{processorName: processorName, registry: registry}
}

func (n ProcessorArtifactRegistry) Add(artifactID ArtifactID, metadata map[string]any) error {
	return n.registry.Add(n.processorName, ArtifactRegistryEntry{
		ArtifactID: artifactID,
		Metadata:   metadata,
	})
}

func (n ProcessorArtifactRegistry) Remove(artifactID ArtifactID) error {
	return n.registry.Remove(n.processorName, artifactID)
}

func (n ProcessorArtifactRegistry) FindByID(artifactID ArtifactID) (_ ArtifactRegistryEntry, _ bool, _ error) {
	return n.registry.FindByID(n.processorName, artifactID)
}

func (n ProcessorArtifactRegistry) FindAll() ([]ArtifactRegistryEntry, error) {
	return n.registry.FindAll(n.processorName)
}

type ArtifactProcessingContext struct {
	context.Context
	Specifications SpecificationGroup
	Artifacts      []Artifact
	Logger         Logger

	ArtifactRegistry ProcessorArtifactRegistry
	processorName    string
}

var _ ArtifactRegistry = (*InMemoryArtifactRegistry)(nil)

// InMemoryArtifactRegistry maintains a registry in memory.
// It can be useful for tests.
type InMemoryArtifactRegistry struct {
	EntriesMap map[string][]ArtifactRegistryEntry
	mu         sync.RWMutex // Mutex to protect concurrent access
}

func (r *InMemoryArtifactRegistry) Add(processorName string, e ArtifactRegistryEntry) error {
	if processorName == "" {
		return errors.NewWithMessage(errors.InternalErrorCode, "processor name is required")
	}
	if e.ArtifactID == "" {
		return errors.NewWithMessage(errors.InternalErrorCode, "artifact id is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.EntriesMap == nil {
		r.EntriesMap = map[string][]ArtifactRegistryEntry{}
	}

	if _, ok := r.EntriesMap[processorName]; !ok {
		r.EntriesMap[processorName] = make([]ArtifactRegistryEntry, 0)
	}

	for i, entry := range r.EntriesMap[processorName] {
		if entry.ArtifactID == e.ArtifactID {
			r.EntriesMap[processorName] = slices.Delete(r.EntriesMap[processorName], i, i+1)
		}
	}

	r.EntriesMap[processorName] = append(r.EntriesMap[processorName], e)

	return nil
}

func (r *InMemoryArtifactRegistry) Remove(processorName string, artifactID ArtifactID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if processorName == "" {
		return errors.NewWithMessage(errors.InternalErrorCode, "processor name is required")
	}
	if artifactID == "" {
		return errors.NewWithMessage(errors.InternalErrorCode, "artifact id is required")
	}

	if _, ok := r.EntriesMap[processorName]; !ok {
		return nil
	}

	var artifacts []ArtifactRegistryEntry
	for _, entry := range r.EntriesMap[processorName] {
		if entry.ArtifactID != artifactID {
			artifacts = append(artifacts, entry)
		}
	}

	r.EntriesMap[processorName] = artifacts

	return nil
}

func (r *InMemoryArtifactRegistry) FindByID(processorName string, artifactID ArtifactID) (entry ArtifactRegistryEntry, found bool, err error) {
	all, _ := r.FindAll(processorName)

	for _, e := range all {
		if e.ArtifactID == artifactID {
			return e, true, nil
		}
	}

	return ArtifactRegistryEntry{}, false, nil
}

func (r *InMemoryArtifactRegistry) FindAll(processorName string) ([]ArtifactRegistryEntry, error) {
	if r.EntriesMap == nil {
		return nil, nil
	}

	values, ok := r.EntriesMap[processorName]
	if !ok {
		return nil, nil
	}

	return values, nil
}

func (r *InMemoryArtifactRegistry) Load() error { return nil }

func (r *InMemoryArtifactRegistry) Save() error { return nil }

const DefaultJSONArtifactRegistryFileName = ".specter.json"

type JSONArtifactRegistryRepresentation struct {
	GeneratedAt time.Time                              `json:"generatedAt"`
	EntriesMap  map[string][]JSONArtifactRegistryEntry `json:"entries"`
}

type JSONArtifactRegistryEntry struct {
	ArtifactID string         `json:"artifactId"`
	Metadata   map[string]any `json:"metadata"`
}

var _ ArtifactRegistry = (*JSONArtifactRegistry)(nil)

// JSONArtifactRegistry implementation of a ArtifactRegistry that is saved as a JSON file.
type JSONArtifactRegistry struct {
	*InMemoryArtifactRegistry
	FileSystem   FileSystem
	FilePath     string
	TimeProvider TimeProvider

	mu sync.RWMutex // Mutex to protect concurrent access
}

func (r *JSONArtifactRegistry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	bytes, err := r.FileSystem.ReadFile(r.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading artifact file registry")
	}

	// empty file is okay
	if len(bytes) == 0 {
		return nil
	}

	repr := &JSONArtifactRegistryRepresentation{}

	if err := json.Unmarshal(bytes, repr); err != nil {
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading artifact file registry")
	}

	for processorName, entries := range repr.EntriesMap {
		for _, entry := range entries {
			if err := r.InMemoryArtifactRegistry.Add(processorName, ArtifactRegistryEntry{
				ArtifactID: ArtifactID(entry.ArtifactID),
				Metadata:   entry.Metadata,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *JSONArtifactRegistry) Save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repr := JSONArtifactRegistryRepresentation{
		GeneratedAt: r.TimeProvider(),
		EntriesMap:  make(map[string][]JSONArtifactRegistryEntry, len(r.InMemoryArtifactRegistry.EntriesMap)),
	}

	// Add entries to representation
	for processorName, entries := range r.InMemoryArtifactRegistry.EntriesMap {
		repr.EntriesMap[processorName] = nil
		for _, entry := range entries {
			repr.EntriesMap[processorName] = append(repr.EntriesMap[processorName], JSONArtifactRegistryEntry{
				ArtifactID: string(entry.ArtifactID),
				Metadata:   entry.Metadata,
			})
		}
	}

	// Set generation date
	repr.GeneratedAt = r.TimeProvider()

	// Generate a JSON file containing all artifact files for clean up later on
	js, err := json.MarshalIndent(repr, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed generating artifact file registry")
	}
	if err := r.FileSystem.WriteFile(r.FilePath, js, fs.ModePerm); err != nil {
		return errors.Wrap(err, "failed generating artifact file registry")
	}

	return nil
}
