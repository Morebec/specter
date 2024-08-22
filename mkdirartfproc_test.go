package specter

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMakeDirectoryArtifactsProcessor_Process(t *testing.T) {
	tests := []struct {
		name         string
		mockFS       *mockFileSystem
		artifacts    []Artifact
		expectedDirs []string
		expectError  error
	}{
		{
			name: "GIVEN directories THEN successful directory creation",
			mockFS: &mockFileSystem{
				dirs: make(map[string]bool),
			},
			artifacts: []Artifact{
				{
					Name:  "dir1",
					Value: DirectoryArtifact{Path: "/path/to/dir1", Mode: 0755},
				},
				{
					Name:  "dir2",
					Value: DirectoryArtifact{Path: "/path/to/dir2", Mode: 0755},
				},
			},
			expectedDirs: []string{"/path/to/dir1", "/path/to/dir2"},
			expectError:  nil,
		},
		{
			name: "GIVEN non-directory artifacts THEN skip and return no error",
			mockFS: &mockFileSystem{
				dirs: make(map[string]bool),
			},
			artifacts: []Artifact{
				{
					Name:  "dir1",
					Value: DirectoryArtifact{Path: "/path/to/dir1", Mode: 0755},
				},
				{
					Name:  "not_a_dir",
					Value: "this is not a directory",
				},
			},
			expectedDirs: []string{"/path/to/dir1"},
			expectError:  nil,
		},
		{
			name: "GIVEN error creating directory THEN return an error",
			mockFS: &mockFileSystem{
				dirs:     make(map[string]bool),
				mkdirErr: assert.AnError,
			},
			artifacts: []Artifact{
				{
					Name:  "dir1",
					Value: DirectoryArtifact{Path: "/path/to/dir1", Mode: 0755},
				},
			},
			expectedDirs: []string{},
			expectError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := MakeDirectoryArtifactsProcessor{FileSystem: tt.mockFS}
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

			for _, dir := range tt.expectedDirs {
				assert.True(t, tt.mockFS.dirs[dir], "directory %q should have been created", dir)
			}
		})
	}
}

func TestMakeDirectoryArtifactsProcessor_Name(t *testing.T) {
	p := MakeDirectoryArtifactsProcessor{}
	require.NotEqual(t, "", p.Name())
}
