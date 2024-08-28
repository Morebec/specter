package specter

import (
	"encoding/json"
	"github.com/morebec/go-errors/errors"
	"os"
	"sync"
	"time"
)

var _ ArtifactRegistry = (*InMemoryArtifactRegistry)(nil)

type InMemoryArtifactRegistry struct {
	ArtifactMap map[string][]ArtifactRegistryEntry
	mu          sync.RWMutex // Mutex to protect concurrent access
}

func (r *InMemoryArtifactRegistry) Add(processorName string, e ArtifactRegistryEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ArtifactMap == nil {
		r.ArtifactMap = map[string][]ArtifactRegistryEntry{}
	}

	if _, ok := r.ArtifactMap[processorName]; !ok {
		r.ArtifactMap[processorName] = make([]ArtifactRegistryEntry, 0)
	}
	r.ArtifactMap[processorName] = append(r.ArtifactMap[processorName], e)

	return nil
}

func (r *InMemoryArtifactRegistry) Remove(processorName string, artifactID ArtifactID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.ArtifactMap[processorName]; !ok {
		return nil
	}
	if r.ArtifactMap == nil {
		r.ArtifactMap = map[string][]ArtifactRegistryEntry{}
	}

	var artifacts []ArtifactRegistryEntry
	for _, entry := range r.ArtifactMap[processorName] {
		if entry.ArtifactID != artifactID {
			artifacts = append(artifacts, entry)
		}
	}

	r.ArtifactMap[processorName] = artifacts

	return nil
}

func (r *InMemoryArtifactRegistry) FindByID(processorName string, artifactID ArtifactID) (entry ArtifactRegistryEntry, found bool, err error) {
	all, err := r.FindAll(processorName)
	if err != nil {
		return ArtifactRegistryEntry{}, false, err
	}

	for _, e := range all {
		if e.ArtifactID == artifactID {
			return e, true, nil
		}
	}

	return ArtifactRegistryEntry{}, false, nil
}

func (r *InMemoryArtifactRegistry) FindAll(processorName string) ([]ArtifactRegistryEntry, error) {
	if r.ArtifactMap == nil {
		r.ArtifactMap = map[string][]ArtifactRegistryEntry{}
	}

	values, ok := r.ArtifactMap[processorName]
	if !ok {
		return nil, nil
	}

	return values, nil
}

func (r *InMemoryArtifactRegistry) Load() error { return nil }

func (r *InMemoryArtifactRegistry) Save() error { return nil }

type JsonArtifactRegistryEntry struct {
	ArtifactID ArtifactID     `json:"artifactId"`
	Metadata   map[string]any `json:"metadata"`
}

var _ ArtifactRegistry = (*JSONArtifactRegistry)(nil)

// JSONArtifactRegistry implementation of a ArtifactRegistry that is saved as a JSON file.
type JSONArtifactRegistry struct {
	FileSystem          FileSystem       `json:"-"`
	FilePath            string           `json:"-"`
	CurrentTimeProvider func() time.Time `json:"-"`

	GeneratedAt time.Time                              `json:"generatedAt"`
	Entries     map[string][]JsonArtifactRegistryEntry `json:"entries"`
	mu          sync.RWMutex                           // Mutex to protect concurrent access
}

func (r *JSONArtifactRegistry) Add(processorName string, e ArtifactRegistryEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.Entries[processorName]; !ok {
		r.Entries[processorName] = make([]JsonArtifactRegistryEntry, 0)
	}

	r.Entries[processorName] = append(r.Entries[processorName], JsonArtifactRegistryEntry{
		ArtifactID: e.ArtifactID,
		Metadata:   e.Metadata,
	})

	return nil
}

func (r *JSONArtifactRegistry) Remove(processorName string, artifactID ArtifactID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.Entries[processorName]; !ok {
		return nil
	}

	var entries []JsonArtifactRegistryEntry
	for _, entry := range r.Entries[processorName] {
		if entry.ArtifactID != artifactID {
			entries = append(entries, entry)
		}
	}

	r.Entries[processorName] = entries

	return nil
}

func (r *JSONArtifactRegistry) FindByID(processorName string, artifactID ArtifactID) (ArtifactRegistryEntry, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.Entries[processorName]; !ok {
		return ArtifactRegistryEntry{}, false, nil
	}

	for _, entry := range r.Entries[processorName] {
		if entry.ArtifactID == artifactID {
			return ArtifactRegistryEntry{
				ArtifactID: entry.ArtifactID,
				Metadata:   entry.Metadata,
			}, true, nil
		}
	}

	return ArtifactRegistryEntry{}, false, nil
}

func (r *JSONArtifactRegistry) FindAll(processorName string) ([]ArtifactRegistryEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.Entries[processorName]; !ok {
		return nil, nil
	}

	var entries []ArtifactRegistryEntry
	for _, entry := range r.Entries[processorName] {
		entries = append(entries, ArtifactRegistryEntry{
			ArtifactID: entry.ArtifactID,
			Metadata:   entry.Metadata,
		})
	}

	return entries, nil
}

// NewJSONArtifactRegistry returns a new artifact file registry.
func NewJSONArtifactRegistry(fileName string, fs FileSystem) *JSONArtifactRegistry {
	return &JSONArtifactRegistry{
		Entries:  make(map[string][]JsonArtifactRegistryEntry),
		FilePath: fileName,
		CurrentTimeProvider: func() time.Time {
			return time.Now()
		},
		FileSystem: fs,
	}
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

	if err := json.Unmarshal(bytes, r); err != nil {
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading artifact file registry")
	}

	return nil
}

func (r *JSONArtifactRegistry) Save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.GeneratedAt = r.CurrentTimeProvider()

	// Generate a JSON file containing all artifact files for clean up later on
	js, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed generating artifact file registry")
	}
	if err := r.FileSystem.WriteFile(r.FilePath, js, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed generating artifact file registry")
	}

	return nil
}
