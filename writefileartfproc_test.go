package specter

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteFileArtifactProcessor_Process(t *testing.T) {
	tests := []struct {
		name          string
		mockFS        *mockFileSystem
		artifacts     []Artifact
		expectedFiles []string
		expectError   error
	}{
		{
			name:   "GIVEN file artifacts THEN successful file creation",
			mockFS: &mockFileSystem{},
			artifacts: []Artifact{
				{
					Name:  "file1",
					Value: FileArtifact{Path: "/path/to/file1", Mode: 0755},
				},
				{
					Name:  "file2",
					Value: FileArtifact{Path: "/path/to/file2", Mode: 0755},
				},
			},
			expectedFiles: []string{"/path/to/file1", "/path/to/file2"},
			expectError:   nil,
		},
		{
			name:   "GIVEN non-file artifacts THEN skip and return no error",
			mockFS: &mockFileSystem{},
			artifacts: []Artifact{
				{
					Name:  "file1",
					Value: FileArtifact{Path: "/path/to/file1", Mode: 0755},
				},
				{
					Name:  "not_a_file",
					Value: "this is not a file",
				},
			},
			expectedFiles: []string{"/path/to/file1"},
			expectError:   nil,
		},
		{
			name: "GIVEN error creating file THEN return an error",
			mockFS: &mockFileSystem{
				writeFileErr: assert.AnError,
			},
			artifacts: []Artifact{
				{
					Name:  "file1",
					Value: FileArtifact{Path: "/path/to/file1", Mode: 0755},
				},
			},
			expectedFiles: []string{},
			expectError:   assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := WriteFileArtifactProcessor{FileSystem: tt.mockFS}
			ctx := ArtifactProcessingContext{
				Context:          context.Background(),
				Artifacts:        tt.artifacts,
				Logger:           NewDefaultLogger(DefaultLoggerConfig{}),
				artifactRegistry: &InMemoryArtifactRegistry{},
				processorName:    processor.Name(),
			}
			err := processor.Process(ctx)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.expectError.Error())
			} else {
				assert.NoError(t, err)
			}

			for _, file := range tt.expectedFiles {
				_, fileExists := tt.mockFS.files[file]
				assert.True(t, fileExists, "file %q should have been created", file)
			}
		})
	}
}

func TestWriteFileArtifactProcessor_Name(t *testing.T) {
	assert.NotEqual(t, "", WriteFileArtifactProcessor{}.Name())
}
