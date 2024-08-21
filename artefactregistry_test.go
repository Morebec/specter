package specter

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJSONArtifactRegistry_Load(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   string
		expectError   bool
		expectedValue *JSONArtifactRegistry
	}{
		{
			name:        "Successful Load",
			fileContent: `{"generatedAt":"2024-01-01T00:00:00Z","files":{"processor1":{"files":["file1.txt"]}}}`,
			expectError: false,
			expectedValue: &JSONArtifactRegistry{
				GeneratedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				ArtifactMap: map[string]*JSONArtifactRegistryProcessor{"processor1": {Artifacts: []string{"file1.txt"}}},
			},
		},
		{
			name:        "File Not Exist",
			fileContent: "",
			expectError: false,
			expectedValue: &JSONArtifactRegistry{
				GeneratedAt: time.Time{},
				ArtifactMap: nil,
			},
		},
		{
			name:        "Malformed JSON",
			fileContent: `{"files":{`,
			expectError: true,
			expectedValue: &JSONArtifactRegistry{
				GeneratedAt: time.Time{},
				ArtifactMap: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			filePath := "test_registry.json"
			err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}
			defer os.Remove(filePath)

			registry := &JSONArtifactRegistry{
				FilePath: filePath,
			}

			// Act
			err = registry.Load()

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, registry)
			}
		})
	}
}

func TestJSONArtifactRegistry_Save(t *testing.T) {
	tests := []struct {
		name         string
		initialState *JSONArtifactRegistry
		expectedJSON string
	}{
		{
			name: "Successful Save",
			initialState: &JSONArtifactRegistry{
				ArtifactMap: map[string]*JSONArtifactRegistryProcessor{
					"processor1": {
						Artifacts: []string{"file1.txt"},
					},
				},
			},
			expectedJSON: `{
  "generatedAt": "0001-01-01T00:00:00Z",
  "files": {
    "processor1": {
      "files": [
        "file1.txt"
      ]
    }
  }
}`,
		},
		{
			name:         "Empty Registry",
			initialState: &JSONArtifactRegistry{},
			expectedJSON: `{
  "generatedAt": "0001-01-01T00:00:00Z",
  "files": {}
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			filePath := "test_registry.json"
			registry := &JSONArtifactRegistry{
				FilePath: filePath,
			}
			registry.ArtifactMap = tt.initialState.ArtifactMap

			// Act
			err := registry.Save()

			// Assert
			assert.NoError(t, err)

			// Read back and verify
			data, err := os.ReadFile(filePath)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expectedJSON, string(data))
		})
	}
}

func TestJSONArtifactRegistry_AddArtifact(t *testing.T) {
	tests := []struct {
		name          string
		initialMap    map[string]*JSONArtifactRegistryProcessor
		processorName string
		artifactName  string
		expectedMap   map[string]*JSONArtifactRegistryProcessor
	}{
		{
			name:          "Add New Artifact",
			initialMap:    map[string]*JSONArtifactRegistryProcessor{},
			processorName: "processor1",
			artifactName:  "file1.txt",
			expectedMap: map[string]*JSONArtifactRegistryProcessor{
				"processor1": {
					Artifacts: []string{"file1.txt"},
				},
			},
		},
		{
			name: "Add Artifact to Existing Processor",
			initialMap: map[string]*JSONArtifactRegistryProcessor{
				"processor1": {
					Artifacts: []string{"file2.txt"},
				},
			},
			processorName: "processor1",
			artifactName:  "file1.txt",
			expectedMap: map[string]*JSONArtifactRegistryProcessor{
				"processor1": {
					Artifacts: []string{"file2.txt", "file1.txt"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &JSONArtifactRegistry{
				ArtifactMap: tt.initialMap,
			}

			// Act
			registry.AddArtifact(tt.processorName, tt.artifactName)

			// Assert
			assert.Equal(t, tt.expectedMap, registry.ArtifactMap)
		})
	}
}

func TestJSONArtifactRegistry_RemoveArtifact(t *testing.T) {
	tests := []struct {
		name          string
		initialMap    map[string]*JSONArtifactRegistryProcessor
		processorName string
		artifactName  string
		expectedMap   map[string]*JSONArtifactRegistryProcessor
	}{
		{
			name: "Remove Existing Artifact",
			initialMap: map[string]*JSONArtifactRegistryProcessor{
				"processor1": {
					Artifacts: []string{"file1.txt", "file2.txt"},
				},
			},
			processorName: "processor1",
			artifactName:  "file1.txt",
			expectedMap: map[string]*JSONArtifactRegistryProcessor{
				"processor1": {
					Artifacts: []string{"file2.txt"},
				},
			},
		},
		{
			name: "Remove Non-Existing Artifact",
			initialMap: map[string]*JSONArtifactRegistryProcessor{
				"processor1": {
					Artifacts: []string{"file1.txt"},
				},
			},
			processorName: "processor1",
			artifactName:  "file2.txt",
			expectedMap: map[string]*JSONArtifactRegistryProcessor{
				"processor1": {
					Artifacts: []string{"file1.txt"},
				},
			},
		},
		{
			name: "Remove From Non-Existing Processor",
			initialMap: map[string]*JSONArtifactRegistryProcessor{
				"processor1": {
					Artifacts: []string{"file1.txt"},
				},
			},
			processorName: "processor2",
			artifactName:  "file1.txt",
			expectedMap: map[string]*JSONArtifactRegistryProcessor{
				"processor1": {
					Artifacts: []string{"file1.txt"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &JSONArtifactRegistry{
				ArtifactMap: tt.initialMap,
			}

			// Act
			registry.RemoveArtifact(tt.processorName, tt.artifactName)

			// Assert
			assert.Equal(t, tt.expectedMap, registry.ArtifactMap)
		})
	}
}

func TestJSONArtifactRegistry_Artifacts(t *testing.T) {
	tests := []struct {
		name              string
		initialMap        map[string]*JSONArtifactRegistryProcessor
		processorName     string
		expectedArtifacts []string
	}{
		{
			name: "Get Artifacts for Existing Processor",
			initialMap: map[string]*JSONArtifactRegistryProcessor{
				"processor1": {
					Artifacts: []string{"file1.txt", "file2.txt"},
				},
			},
			processorName:     "processor1",
			expectedArtifacts: []string{"file1.txt", "file2.txt"},
		},
		{
			name: "Get Artifacts for Non-Existing Processor",
			initialMap: map[string]*JSONArtifactRegistryProcessor{
				"processor1": {
					Artifacts: []string{"file1.txt"},
				},
			},
			processorName:     "processor2",
			expectedArtifacts: nil,
		},
		{
			name:              "Empty Registry",
			initialMap:        map[string]*JSONArtifactRegistryProcessor{},
			processorName:     "processor1",
			expectedArtifacts: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &JSONArtifactRegistry{
				ArtifactMap: tt.initialMap,
			}

			// Act
			artifacts := registry.Artifacts(tt.processorName)

			// Assert
			assert.Equal(t, tt.expectedArtifacts, artifacts)
		})
	}
}
