// Copyright 2024 Morébec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package specter

import (
	"encoding/json"
	"github.com/morebec/go-errors/errors"
	"io/fs"
	"os"
	"sync"
	"time"
)

const DefaultJSONArtifactRegistryFileName = ".specter.json"

type JSONArtifactRegistryRepresentation struct {
	GeneratedAt time.Time                              `json:"generatedAt"`
	EntriesMap  map[string][]JSONArtifactRegistryEntry `json:"entries"`
}

type JSONArtifactRegistryEntry struct {
	ArtifactID string         `json:"artifactId"`
	Metadata   map[string]any `json:"metadata"`
}

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
