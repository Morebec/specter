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
	"os"
	"testing"
	"time"
)

func TestJSONArtifactRegistry_InterfaceCompliance(t *testing.T) {
	// JSON Artifact Registry
	assertArtifactRegistryCompliance(t, "JSONArtifactRegistry", func() *JSONArtifactRegistry {
		return NewJSONArtifactRegistry(DefaultJSONArtifactRegistryFileName, &mockFileSystem{})
	})
}

func TestJSONArtifactRegistry_Load(t *testing.T) {
	type given struct {
		jsonFileContent string
		fileSystem      FileSystem
	}
	type then struct {
		expectedError   require.ErrorAssertionFunc
		expectedEntries map[string][]ArtifactRegistryEntry
	}
	tests := []struct {
		name  string
		given given
		then  then
	}{
		{
			name: "Given entries in json file Then entries should be loaded Successfully",
			given: given{
				fileSystem: &mockFileSystem{},
				jsonFileContent: `{
  "generatedAt" : "2024-01-01T00:00:00Z",
  "entries" : {
    "processor1" : [ {
      "artifactId" : "file1.txt",
      "metadata" : null
    } ]
  }
}`,
			},
			then: then{
				expectedError: nil,
				expectedEntries: map[string][]ArtifactRegistryEntry{
					"processor1": {
						{
							ArtifactID: "file1.txt",
						},
					},
				},
			},
		},
		{
			name: "Given file content is empty Then no entries should be loaded",
			given: given{
				fileSystem:      &mockFileSystem{},
				jsonFileContent: "",
			},
			then: then{
				expectedError:   require.NoError,
				expectedEntries: nil,
			},
		},
		{
			name: "Given JSON is malformed Then no entries should be loaded and an error should be returned",
			given: given{
				fileSystem:      &mockFileSystem{},
				jsonFileContent: `{"entries":{`,
			},
			then: then{
				expectedError:   require.Error,
				expectedEntries: nil,
			},
		},
		{
			name: "Given json contains invalid entries Then entries should be loaded Successfully",
			// Invalid entries being processors without a name or artifacts without an ID
			given: given{
				fileSystem: &mockFileSystem{},
				jsonFileContent: `{
  "generatedAt" : "2024-01-01T00:00:00Z",
  "entries" : {
    "processor1" : [ {
      "artifactId" : "",
      "metadata" : null
    } ]
  }
}`,
			},
			then: then{
				expectedError: require.Error,
			},
		},
		{
			name: "Given file does not exist Then return no error",
			given: given{
				fileSystem: &mockFileSystem{
					readFileErr: os.ErrNotExist,
				},
				jsonFileContent: "",
			},
			then: then{
				expectedError:   require.NoError,
				expectedEntries: nil,
			},
		},
		{
			name: "Given filesystem returns an error when reading files Then return an error",
			given: given{
				fileSystem: &mockFileSystem{
					readFileErr: assert.AnError,
				},
				jsonFileContent: "",
			},
			then: then{
				expectedError:   require.Error,
				expectedEntries: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := "test_registry.json"

			fs := tt.given.fileSystem

			err := fs.WriteFile(filePath, []byte(tt.given.jsonFileContent), os.ModePerm)
			require.NoError(t, err)
			defer func(fs FileSystem, path string) {
				require.NoError(t, fs.Remove(path))
			}(fs, filePath)

			registry := NewJSONArtifactRegistry(filePath, fs)
			registry.TimeProvider = staticTimeProvider(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

			err = registry.Load()

			if tt.then.expectedError != nil {
				tt.then.expectedError(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.then.expectedEntries, registry.EntriesMap)
		})
	}
}

func TestJSONArtifactRegistry_Save(t *testing.T) {
	type given struct {
		entries    map[string][]ArtifactRegistryEntry
		fileSystem FileSystem
	}

	type then struct {
		expectedJSON  string
		expectedError require.ErrorAssertionFunc
	}

	tests := []struct {
		name  string
		given given
		then  then
	}{
		{
			name: "GIVEN a set of entries THEN registry should be saved as JSON",
			given: given{
				fileSystem: &mockFileSystem{},
				entries: map[string][]ArtifactRegistryEntry{
					"processor1": {
						{ArtifactID: "file1.txt"},
					},
				},
			},
			then: then{
				expectedJSON: `{
  "generatedAt" : "2024-01-01T00:00:00Z",
  "entries" : {
    "processor1" : [ {
      "artifactId" : "file1.txt",
      "metadata" : null
    } ]
  }
}`,
				expectedError: require.NoError,
			},
		},
		{
			name: "GIVEN no entries THEN registry should be saved as JSON successfully",
			given: given{
				fileSystem: &mockFileSystem{},
				entries:    nil,
			},
			then: then{
				expectedJSON: `{
  "generatedAt": "2024-01-01T00:00:00Z",
  "entries": {}
}`,
				expectedError: require.NoError,
			},
		},
		{
			name: "GIVEN filesystem has errors writing files THEN an error should be returned",
			given: given{
				fileSystem: &mockFileSystem{
					writeFileErr: assert.AnError,
				},
				entries: nil,
			},
			then: then{
				expectedJSON:  "",
				expectedError: require.Error,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.given.fileSystem

			registryFilePath := DefaultJSONArtifactRegistryFileName

			registry := NewJSONArtifactRegistry(registryFilePath, fs)
			registry.TimeProvider = staticTimeProvider(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

			// Add initial entries
			for p, entries := range tt.given.entries {
				for _, e := range entries {
					err := registry.Add(p, e)
					require.NoError(t, err)
				}
			}

			err := registry.Save()
			if tt.then.expectedError != nil {
				tt.then.expectedError(t, err)
			} else {
				require.NoError(t, err)
			}

			// Compare content
			actualJSON, err := fs.ReadFile(registryFilePath)
			if tt.then.expectedJSON != "" {
				require.NoError(t, err)
				assert.JSONEq(t, tt.then.expectedJSON, string(actualJSON))
			} else {
				assert.Equal(t, tt.then.expectedJSON, string(actualJSON))
			}
		})
	}
}
