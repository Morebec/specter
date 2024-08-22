package specter

import (
	"encoding/json"
	"github.com/morebec/go-errors/errors"
	"os"
	"sync"
	"time"
)

// JSONArtifactRegistry implementation of a ArtifactRegistry that is saved as a JSON file.
type JSONArtifactRegistry struct {
	GeneratedAt         time.Time                                 `json:"generatedAt"`
	ArtifactMap         map[string]*JSONArtifactRegistryProcessor `json:"files"`
	FilePath            string                                    `json:"-"`
	FileSystem          FileSystem                                `json:"-"`
	mu                  sync.RWMutex                              // Mutex to protect concurrent access
	CurrentTimeProvider func() time.Time
}

type JSONArtifactRegistryProcessor struct {
	Artifacts []string `json:"files"`
}

// NewJSONArtifactRegistry returns a new artifact file registry.
func NewJSONArtifactRegistry(fileName string, fs FileSystem) *JSONArtifactRegistry {
	return &JSONArtifactRegistry{
		ArtifactMap: nil,
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

func (r *JSONArtifactRegistry) AddArtifact(processorName string, artifactName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.ArtifactMap[processorName]; !ok {
		r.ArtifactMap[processorName] = &JSONArtifactRegistryProcessor{}
	}
	r.ArtifactMap[processorName].Artifacts = append(r.ArtifactMap[processorName].Artifacts, artifactName)
}

func (r *JSONArtifactRegistry) RemoveArtifact(processorName string, artifactName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.ArtifactMap[processorName]; !ok {
		return
	}

	var files []string
	for _, file := range r.ArtifactMap[processorName].Artifacts {
		if file != artifactName {
			files = append(files, file)
		}
	}

	r.ArtifactMap[processorName].Artifacts = files
}

func (r *JSONArtifactRegistry) Artifacts(processorName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	artifacts, ok := r.ArtifactMap[processorName]
	if !ok {
		return nil
	}
	return artifacts.Artifacts
}
