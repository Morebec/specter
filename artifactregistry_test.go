package specter

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJSONArtifactRegistry_Load(t *testing.T) {
	type given struct {
		jsonFileContent string
	}
	type then struct {
		expectedError error
		expectedValue *JSONArtifactRegistry
	}
	tests := []struct {
		name  string
		given given
		then  then
	}{
		{
			name: "Successful Load",
			given: given{
				jsonFileContent: `{"generatedAt":"2024-01-01T00:00:00Z","files":{"processor1":{"files":["file1.txt"]}}}`,
			},
			then: then{
				expectedError: nil,
				expectedValue: &JSONArtifactRegistry{
					GeneratedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					Entries: map[string][]JsonArtifactRegistryEntry{"processor1": {
						{
							ArtifactID: "file1.txt",
						},
					}},
				},
			},
		},
		{
			name: "File Not Exist",
			given: given{
				jsonFileContent: "",
			},
			then: then{
				expectedError: nil,
				expectedValue: &JSONArtifactRegistry{
					GeneratedAt: time.Time{},
					Entries:     nil,
				},
			},
		},
		{
			name: "Malformed JSON",
			given: given{
				jsonFileContent: `{"files":{`,
			},
			then: then{
				expectedError: fmt.Errorf("failed loading artifact file registry: unexpected end of JSON input"),
				expectedValue: &JSONArtifactRegistry{
					GeneratedAt: time.Time{},
					Entries:     nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := "test_registry.json"

			fs := &mockFileSystem{}
			err := fs.WriteFile(filePath, []byte(tt.given.jsonFileContent), os.ModePerm)
			require.NoError(t, err)

			registry := NewJSONArtifactRegistry(filePath, fs)
			registry.CurrentTimeProvider = func() time.Time {
				return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			}

			err = registry.Load()

			if tt.then.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.then.expectedValue.GeneratedAt, registry.GeneratedAt)
			}
		})
	}
}

func TestJSONArtifactRegistry_Save(t *testing.T) {
	type then struct {
		expectedJSON  string
		expectedError error
	}

	tests := []struct {
		name  string
		given *JSONArtifactRegistry
		then  then
	}{
		{
			name: "Successful Save",
			given: &JSONArtifactRegistry{
				UseAbsolutePaths: false,
				Entries: map[string][]JsonArtifactRegistryEntry{
					"processor1": {
						{ArtifactID: "file1.txt"},
					},
				},
			},
			then: then{
				expectedError: nil,
				expectedJSON: `{
  "generatedAt" : "2024-01-01T00:00:00Z",
  "entries" : {
    "processor1" : [ {
      "artifactId" : "file1.txt",
      "metadata" : null
    } ]
  }
}`,
			},
		},
		{
			name:  "Empty Registry",
			given: &JSONArtifactRegistry{},
			then: then{
				expectedJSON: `{
  "generatedAt": "2024-01-01T00:00:00Z",
  "entries": null
}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			filePath := "test_registry.json"
			fs := &mockFileSystem{}
			registry := NewJSONArtifactRegistry(filePath, fs)
			registry.Entries = tt.given.Entries
			registry.CurrentTimeProvider = func() time.Time {
				return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			}

			err := registry.Save()

			if tt.then.expectedError != nil {
				require.ErrorIs(t, err, tt.then.expectedError)
			} else {
				require.NoError(t, err)
			}
			actualJSON, err := fs.ReadFile(filePath)
			assert.JSONEq(t, tt.then.expectedJSON, string(actualJSON))
		})
	}
}

func TestJSONArtifactRegistry_Add(t *testing.T) {
	tests := []struct {
		name          string
		initialMap    map[string][]JsonArtifactRegistryEntry
		processorName string
		entry         ArtifactRegistryEntry
		expectedMap   map[string][]JsonArtifactRegistryEntry
	}{
		{
			name:          "Add New Artifact",
			initialMap:    map[string][]JsonArtifactRegistryEntry{},
			processorName: "processor1",
			entry:         ArtifactRegistryEntry{ArtifactID: "file1.txt"},
			expectedMap: map[string][]JsonArtifactRegistryEntry{
				"processor1": {
					{ArtifactID: "file1.txt"},
				},
			},
		},
		{
			name: "Add Artifact to Existing Processor",
			initialMap: map[string][]JsonArtifactRegistryEntry{
				"processor1": {
					{ArtifactID: "file2.txt"},
				},
			},
			processorName: "processor1",
			entry:         ArtifactRegistryEntry{ArtifactID: "file1.txt"},
			expectedMap: map[string][]JsonArtifactRegistryEntry{
				"processor1": {
					{ArtifactID: "file2.txt"},
					{ArtifactID: "file1.txt"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &JSONArtifactRegistry{
				Entries: tt.initialMap,
			}
			err := registry.Add(tt.processorName, tt.entry)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMap, registry.Entries)
		})
	}
}

func TestJSONArtifactRegistry_Remove(t *testing.T) {
	tests := []struct {
		name          string
		initialMap    map[string][]JsonArtifactRegistryEntry
		processorName string
		artifactID    ArtifactID
		expectedMap   map[string][]JsonArtifactRegistryEntry
	}{
		{
			name: "Remove Existing Artifact",
			initialMap: map[string][]JsonArtifactRegistryEntry{
				"processor1": {
					{ArtifactID: "file1.txt"},
					{ArtifactID: "file2.txt"},
				},
			},
			processorName: "processor1",
			artifactID:    "file1.txt",
			expectedMap: map[string][]JsonArtifactRegistryEntry{
				"processor1": {
					{ArtifactID: "file2.txt"},
				},
			},
		},
		{
			name: "Remove Non-Existing Artifact",
			initialMap: map[string][]JsonArtifactRegistryEntry{
				"processor1": {
					{ArtifactID: "file1.txt"},
				},
			},
			processorName: "processor1",
			artifactID:    "file2.txt",
			expectedMap: map[string][]JsonArtifactRegistryEntry{
				"processor1": {
					{ArtifactID: "file1.txt"},
				},
			},
		},
		{
			name: "Remove From Non-Existing Processor",
			initialMap: map[string][]JsonArtifactRegistryEntry{
				"processor1": {
					{ArtifactID: "file1.txt"},
				},
			},
			processorName: "processor2",
			artifactID:    "file1.txt",
			expectedMap: map[string][]JsonArtifactRegistryEntry{
				"processor1": {
					{ArtifactID: "file1.txt"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &JSONArtifactRegistry{
				Entries: tt.initialMap,
			}
			err := registry.Remove(tt.processorName, tt.artifactID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMap, registry.Entries)
		})
	}
}

func TestJSONArtifactRegistry_FindAll(t *testing.T) {
	tests := []struct {
		name            string
		initialMap      map[string][]JsonArtifactRegistryEntry
		processorName   string
		expectedEntries []ArtifactRegistryEntry
	}{
		{
			name: "FindAll for Existing Processor",
			initialMap: map[string][]JsonArtifactRegistryEntry{
				"processor1": {
					{ArtifactID: "file1.txt"},
					{ArtifactID: "file2.txt"},
				},
			},
			processorName: "processor1",
			expectedEntries: []ArtifactRegistryEntry{
				{ArtifactID: "file1.txt"},
				{ArtifactID: "file2.txt"},
			},
		},
		{
			name: "FindAll for Non-Existing Processor",
			initialMap: map[string][]JsonArtifactRegistryEntry{
				"processor1": {
					{ArtifactID: "file1.txt"},
				},
			},
			processorName:   "processor2",
			expectedEntries: nil,
		},
		{
			name:            "Empty Registry",
			initialMap:      map[string][]JsonArtifactRegistryEntry{},
			processorName:   "processor1",
			expectedEntries: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &JSONArtifactRegistry{
				Entries: tt.initialMap,
			}

			artifacts, err := registry.FindAll(tt.processorName)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedEntries, artifacts)
		})
	}
}
