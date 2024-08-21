package specter

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJSONOutputRegistry_Load(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   string
		expectError   bool
		expectedValue *JSONOutputRegistry
	}{
		{
			name:        "Successful Load",
			fileContent: `{"generatedAt":"2024-01-01T00:00:00Z","files":{"processor1":{"files":["file1.txt"]}}}`,
			expectError: false,
			expectedValue: &JSONOutputRegistry{
				GeneratedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				OutputMap:   map[string]*JSONOutputRegistryProcessor{"processor1": {Outputs: []string{"file1.txt"}}},
			},
		},
		{
			name:        "File Not Exist",
			fileContent: "",
			expectError: false,
			expectedValue: &JSONOutputRegistry{
				GeneratedAt: time.Time{},
				OutputMap:   nil,
			},
		},
		{
			name:        "Malformed JSON",
			fileContent: `{"files":{`,
			expectError: true,
			expectedValue: &JSONOutputRegistry{
				GeneratedAt: time.Time{},
				OutputMap:   nil,
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

			registry := &JSONOutputRegistry{
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

func TestJSONOutputRegistry_Save(t *testing.T) {
	tests := []struct {
		name         string
		initialState *JSONOutputRegistry
		expectedJSON string
	}{
		{
			name: "Successful Save",
			initialState: &JSONOutputRegistry{
				OutputMap: map[string]*JSONOutputRegistryProcessor{
					"processor1": {
						Outputs: []string{"file1.txt"},
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
			initialState: &JSONOutputRegistry{},
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
			registry := &JSONOutputRegistry{
				FilePath: filePath,
			}
			registry.OutputMap = tt.initialState.OutputMap

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

func TestJSONOutputRegistry_AddOutput(t *testing.T) {
	tests := []struct {
		name          string
		initialMap    map[string]*JSONOutputRegistryProcessor
		processorName string
		outputName    string
		expectedMap   map[string]*JSONOutputRegistryProcessor
	}{
		{
			name:          "Add New Output",
			initialMap:    map[string]*JSONOutputRegistryProcessor{},
			processorName: "processor1",
			outputName:    "file1.txt",
			expectedMap: map[string]*JSONOutputRegistryProcessor{
				"processor1": {
					Outputs: []string{"file1.txt"},
				},
			},
		},
		{
			name: "Add Output to Existing Processor",
			initialMap: map[string]*JSONOutputRegistryProcessor{
				"processor1": {
					Outputs: []string{"file2.txt"},
				},
			},
			processorName: "processor1",
			outputName:    "file1.txt",
			expectedMap: map[string]*JSONOutputRegistryProcessor{
				"processor1": {
					Outputs: []string{"file2.txt", "file1.txt"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &JSONOutputRegistry{
				OutputMap: tt.initialMap,
			}

			// Act
			registry.AddOutput(tt.processorName, tt.outputName)

			// Assert
			assert.Equal(t, tt.expectedMap, registry.OutputMap)
		})
	}
}

func TestJSONOutputRegistry_RemoveOutput(t *testing.T) {
	tests := []struct {
		name          string
		initialMap    map[string]*JSONOutputRegistryProcessor
		processorName string
		outputName    string
		expectedMap   map[string]*JSONOutputRegistryProcessor
	}{
		{
			name: "Remove Existing Output",
			initialMap: map[string]*JSONOutputRegistryProcessor{
				"processor1": {
					Outputs: []string{"file1.txt", "file2.txt"},
				},
			},
			processorName: "processor1",
			outputName:    "file1.txt",
			expectedMap: map[string]*JSONOutputRegistryProcessor{
				"processor1": {
					Outputs: []string{"file2.txt"},
				},
			},
		},
		{
			name: "Remove Non-Existing Output",
			initialMap: map[string]*JSONOutputRegistryProcessor{
				"processor1": {
					Outputs: []string{"file1.txt"},
				},
			},
			processorName: "processor1",
			outputName:    "file2.txt",
			expectedMap: map[string]*JSONOutputRegistryProcessor{
				"processor1": {
					Outputs: []string{"file1.txt"},
				},
			},
		},
		{
			name: "Remove From Non-Existing Processor",
			initialMap: map[string]*JSONOutputRegistryProcessor{
				"processor1": {
					Outputs: []string{"file1.txt"},
				},
			},
			processorName: "processor2",
			outputName:    "file1.txt",
			expectedMap: map[string]*JSONOutputRegistryProcessor{
				"processor1": {
					Outputs: []string{"file1.txt"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &JSONOutputRegistry{
				OutputMap: tt.initialMap,
			}

			// Act
			registry.RemoveOutput(tt.processorName, tt.outputName)

			// Assert
			assert.Equal(t, tt.expectedMap, registry.OutputMap)
		})
	}
}

func TestJSONOutputRegistry_Outputs(t *testing.T) {
	tests := []struct {
		name            string
		initialMap      map[string]*JSONOutputRegistryProcessor
		processorName   string
		expectedOutputs []string
	}{
		{
			name: "Get Outputs for Existing Processor",
			initialMap: map[string]*JSONOutputRegistryProcessor{
				"processor1": {
					Outputs: []string{"file1.txt", "file2.txt"},
				},
			},
			processorName:   "processor1",
			expectedOutputs: []string{"file1.txt", "file2.txt"},
		},
		{
			name: "Get Outputs for Non-Existing Processor",
			initialMap: map[string]*JSONOutputRegistryProcessor{
				"processor1": {
					Outputs: []string{"file1.txt"},
				},
			},
			processorName:   "processor2",
			expectedOutputs: nil,
		},
		{
			name:            "Empty Registry",
			initialMap:      map[string]*JSONOutputRegistryProcessor{},
			processorName:   "processor1",
			expectedOutputs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &JSONOutputRegistry{
				OutputMap: tt.initialMap,
			}

			// Act
			outputs := registry.Outputs(tt.processorName)

			// Assert
			assert.Equal(t, tt.expectedOutputs, outputs)
		})
	}
}
