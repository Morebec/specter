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
	ArtifactMap map[string][]ArtifactID
	mu          sync.RWMutex // Mutex to protect concurrent access
}

func (r *InMemoryArtifactRegistry) Load() error { return nil }

func (r *InMemoryArtifactRegistry) Save() error { return nil }

func (r *InMemoryArtifactRegistry) AddArtifact(processorName string, artifactID ArtifactID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ArtifactMap == nil {
		r.ArtifactMap = map[string][]ArtifactID{}
	}

	if _, ok := r.ArtifactMap[processorName]; !ok {
		r.ArtifactMap[processorName] = make([]ArtifactID, 0)
	}
	r.ArtifactMap[processorName] = append(r.ArtifactMap[processorName], artifactID)
}

func (r *InMemoryArtifactRegistry) RemoveArtifact(processorName string, artifactID ArtifactID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.ArtifactMap[processorName]; !ok {
		return
	}
	if r.ArtifactMap == nil {
		r.ArtifactMap = map[string][]ArtifactID{}
	}

	var artifacts []ArtifactID
	for _, file := range r.ArtifactMap[processorName] {
		if file != artifactID {
			artifacts = append(artifacts, file)
		}
	}

	r.ArtifactMap[processorName] = artifacts
}

func (r *InMemoryArtifactRegistry) Artifacts(processorName string) []ArtifactID {
	if r.ArtifactMap == nil {
		r.ArtifactMap = map[string][]ArtifactID{}
	}

	values, ok := r.ArtifactMap[processorName]
	if !ok {
		return nil
	}

	return values
}

// JSONArtifactRegistry implementation of a ArtifactRegistry that is saved as a JSON file.
type JSONArtifactRegistry struct {
	UseAbsolutePaths bool `json:"-"`

	GeneratedAt         time.Time                                 `json:"generatedAt"`
	ArtifactMap         map[string]*JSONArtifactRegistryProcessor `json:"files"`
	FilePath            string                                    `json:"-"`
	FileSystem          FileSystem                                `json:"-"`
	mu                  sync.RWMutex                              // Mutex to protect concurrent access
	CurrentTimeProvider func() time.Time                          `json:"-"`
}

type JSONArtifactRegistryProcessor struct {
	Artifacts []ArtifactID `json:"files"`
}

// NewJSONArtifactRegistry returns a new artifact file registry.
func NewJSONArtifactRegistry(fileName string, fs FileSystem) *JSONArtifactRegistry {
	return &JSONArtifactRegistry{
		ArtifactMap: map[string]*JSONArtifactRegistryProcessor{},
		FilePath:    fileName,
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

func (r *JSONArtifactRegistry) AddArtifact(processorName string, artifactID ArtifactID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ArtifactMap == nil {
		r.ArtifactMap = map[string]*JSONArtifactRegistryProcessor{}
	}

	if _, ok := r.ArtifactMap[processorName]; !ok {
		r.ArtifactMap[processorName] = &JSONArtifactRegistryProcessor{}
	}
	r.ArtifactMap[processorName].Artifacts = append(r.ArtifactMap[processorName].Artifacts, artifactID)
}

func (r *JSONArtifactRegistry) RemoveArtifact(processorName string, artifactID ArtifactID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.ArtifactMap[processorName]; !ok {
		return
	}

	var files []ArtifactID
	for _, file := range r.ArtifactMap[processorName].Artifacts {
		if file != artifactID {
			files = append(files, file)
		}
	}

	r.ArtifactMap[processorName].Artifacts = files
}

func (r *JSONArtifactRegistry) Artifacts(processorName string) []ArtifactID {
	r.mu.RLock()
	defer r.mu.RUnlock()

	artifacts, ok := r.ArtifactMap[processorName]
	if !ok {
		return nil
	}
	return artifacts.Artifacts
}
