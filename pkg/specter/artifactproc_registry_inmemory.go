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
	"github.com/morebec/go-errors/errors"
	"slices"
	"sync"
)

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
