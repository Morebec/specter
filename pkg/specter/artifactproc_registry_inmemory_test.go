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

package specter_test

import (
	. "github.com/morebec/specter/pkg/specter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInMemoryArtifactRegistry_InterfaceCompliance(t *testing.T) {
	// InMemoryArtifactRegistry
	assertArtifactRegistryCompliance(t, "InMemoryArtifactRegistry", func() *InMemoryArtifactRegistry {
		return &InMemoryArtifactRegistry{}
	})
}

func TestInMemoryArtifactRegistry_Add(t *testing.T) {
	type given struct {
		EntriesMap map[string][]ArtifactRegistryEntry
	}
	type when struct {
		processorName string
		entry         ArtifactRegistryEntry
	}
	type then struct {
		expectedError   assert.ErrorAssertionFunc
		expectedEntries []ArtifactRegistryEntry
	}
	tests := []struct {
		name  string
		given given
		when  when
		then  then
	}{
		{
			name: "Given non nil entry map When an entry is added Then the entry should be in the registry",
			given: given{
				EntriesMap: map[string][]ArtifactRegistryEntry{},
			},
			when: when{
				processorName: "processor1",
				entry: ArtifactRegistryEntry{
					ArtifactID: "an_artifact",
				},
			},
			then: then{
				expectedEntries: []ArtifactRegistryEntry{
					{
						ArtifactID: "an_artifact",
					},
				},
				expectedError: assert.NoError,
			},
		},
		{
			name: "Given a nil entry map When an entry is added Then the entry should be in the registry",
			given: given{
				EntriesMap: nil,
			},
			when: when{
				processorName: "processor1",
				entry: ArtifactRegistryEntry{
					ArtifactID: "an_artifact",
				},
			},
			then: then{
				expectedEntries: []ArtifactRegistryEntry{
					{
						ArtifactID: "an_artifact",
					},
				},
				expectedError: assert.NoError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &InMemoryArtifactRegistry{
				EntriesMap: tt.given.EntriesMap,
			}
			err := r.Add(tt.when.processorName, tt.when.entry)
			if tt.then.expectedError != nil {
				tt.then.expectedError(t, err)
			}
			if tt.then.expectedEntries != nil {
				require.Equal(t, tt.then.expectedEntries, r.EntriesMap[tt.when.processorName])
			}
		})
	}
}

func TestInMemoryArtifactRegistry_FindAll(t *testing.T) {
	type given struct {
		EntriesMap map[string][]ArtifactRegistryEntry
	}
	type when struct {
		processorName string
	}
	type then struct {
		expectedArtifacts []ArtifactRegistryEntry
		expectedError     assert.ErrorAssertionFunc
	}
	tests := []struct {
		name  string
		given given
		when  when
		then  then
	}{
		{
			name: "Given nil entry map Then return nil",
			given: given{
				EntriesMap: nil,
			},
			when: when{
				processorName: "unit_tester",
			},
			then: then{
				expectedArtifacts: nil,
				expectedError:     assert.NoError,
			},
		},
		{
			name: "Given non nil empty entry map Then return nil",
			given: given{
				EntriesMap: map[string][]ArtifactRegistryEntry{},
			},
			when: when{
				processorName: "unit_tester",
			},
			then: then{
				expectedArtifacts: nil,
				expectedError:     assert.NoError,
			},
		},
		{
			name: "Given nil slice for processor name Then return nil",
			given: given{
				EntriesMap: map[string][]ArtifactRegistryEntry{
					"unit_tester": nil,
				},
			},
			when: when{
				processorName: "unit_tester",
			},
			then: then{
				expectedArtifacts: nil,
				expectedError:     assert.NoError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &InMemoryArtifactRegistry{
				EntriesMap: tt.given.EntriesMap,
			}
			artifacts, err := r.FindAll(tt.when.processorName)
			if tt.then.expectedError != nil {
				tt.then.expectedError(t, err)
			}
			require.Equal(t, tt.then.expectedArtifacts, artifacts)
		})
	}
}

func TestInMemoryArtifactRegistry_Load(t *testing.T) {
	r := &InMemoryArtifactRegistry{}
	require.Nil(t, r.Load())
}

func TestInMemoryArtifactRegistry_Save(t *testing.T) {
	r := &InMemoryArtifactRegistry{}
	require.Nil(t, r.Save())
}

func TestInMemoryArtifactRegistry_Remove(t *testing.T) {
	type given struct {
		EntriesMap map[string][]ArtifactRegistryEntry
	}
	type when struct {
		processorName string
		artifactID    ArtifactID
	}
	type then struct {
		expectedEntries []ArtifactRegistryEntry
		expectedError   assert.ErrorAssertionFunc
	}
	tests := []struct {
		name  string
		given given
		when  when
		then  then
	}{
		{
			name: "Given nil entry map Then return nil",
			given: given{
				EntriesMap: nil,
			},
			when: when{
				processorName: "unit_tester",
				artifactID:    "does_not_exist",
			},
			then: then{
				expectedEntries: nil,
				expectedError:   assert.NoError,
			},
		},
		{
			name: "Given empty entry map Then return nil",
			given: given{
				EntriesMap: map[string][]ArtifactRegistryEntry{},
			},
			when: when{
				processorName: "unit_tester",
				artifactID:    "does_not_exist",
			},
			then: then{
				expectedEntries: nil,
				expectedError:   assert.NoError,
			},
		},
		{
			name: "Given nil slice for processor Then return nil",
			given: given{
				EntriesMap: map[string][]ArtifactRegistryEntry{
					"unit_tester": nil,
				},
			},
			when: when{
				processorName: "unit_tester",
				artifactID:    "does_not_exist",
			},
			then: then{
				expectedEntries: nil,
				expectedError:   assert.NoError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &InMemoryArtifactRegistry{
				EntriesMap: tt.given.EntriesMap,
			}

			err := r.Remove(tt.when.processorName, tt.when.artifactID)
			if tt.then.expectedError != nil {
				tt.then.expectedError(t, err)
			}

			all, err := r.FindAll(tt.when.processorName)
			require.NoError(t, err)
			require.Equal(t, tt.then.expectedEntries, all)
		})
	}
}
