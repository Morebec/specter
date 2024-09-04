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
	"context"
	"github.com/morebec/go-errors/errors"
	"github.com/morebec/specter/pkg/specter"
	"github.com/morebec/specter/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"testing"
	"time"
)

func TestWriteFileArtifactProcessor_Process(t *testing.T) {
	type processGiven struct {
		fileSystem      specter.FileSystem
		registryEntries []specter.ArtifactRegistryEntry
		contextTimeout  time.Duration
	}
	type processWhen struct {
		artifacts []specter.Artifact
	}
	type processThen struct {
		expectedFiles []string
		expectedError require.ErrorAssertionFunc
	}

	tests := []struct {
		name  string
		given processGiven
		when  processWhen
		then  processThen
	}{
		// CONTEXT CANCELLATIONS
		{
			name: "WHEN context cancels Then return context error",
			given: processGiven{
				fileSystem:     &testutils.MockFileSystem{},
				contextTimeout: time.Nanosecond,
			},
			then: processThen{
				expectedError: func(t require.TestingT, err error, i ...interface{}) {
					require.ErrorIs(t, err, context.DeadlineExceeded)
				},
			},
		},

		// CLEAN
		{
			name: "GIVEN registry entries without metadata " +
				"WHEN cleanup fails as a result " +
				"THEN an error should be returned",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{},
				registryEntries: []specter.ArtifactRegistryEntry{
					{
						ArtifactID: "/path/to/file1",
						Metadata:   nil,
					},
				},
			},
			then: processThen{
				expectedError: testutils.RequireErrorWithCode(specter.FileArtifactProcessorCleanUpFailedErrorCode),
			},
		},
		{
			name: "GIVEN registry entries with invalid paths metadata " +
				"WHEN cleanup fails as a result " +
				"THEN an error should be returned",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{},
				registryEntries: []specter.ArtifactRegistryEntry{
					{
						ArtifactID: "/path/to/file1",
						Metadata: map[string]any{
							"path":      "", // no path
							"writeMode": string(specter.WriteOnceMode),
						},
					},
				},
			},
			then: processThen{
				expectedError: testutils.RequireErrorWithCode(specter.FileArtifactProcessorCleanUpFailedErrorCode),
			},
		},
		{
			name: "GIVEN registry entries with invalid writeMode metadata " +
				"WHEN cleanup fails as a result " +
				"THEN an error should be returned",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{},
				registryEntries: []specter.ArtifactRegistryEntry{
					{
						ArtifactID: "/path/to/file1",
						Metadata: map[string]any{
							"path":      "/path/to/file1",
							"writeMode": "", // no writeMode
						},
					},
				},
			},
			then: processThen{
				expectedError: testutils.RequireErrorWithCode(specter.FileArtifactProcessorCleanUpFailedErrorCode),
			},
		},
		{
			name: "GIVEN file system fails to remove files " +
				"WHEN processing cleanup fails as a result" +
				"THEN an error should be returned",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{
					RmErr: assert.AnError,
					Files: map[string][]byte{
						"/path/to/file1": []byte("file content"),
					},
				},
				registryEntries: []specter.ArtifactRegistryEntry{
					{
						ArtifactID: "/path/to/file1",
						Metadata: map[string]interface{}{
							"path":      "/path/to/file1",
							"writeMode": string(specter.RecreateMode),
						},
					},
				},
			},
			when: processWhen{
				artifacts: []specter.Artifact{
					&specter.FileArtifact{Path: "/path/to/file1", FileMode: os.ModePerm, WriteMode: specter.RecreateMode},
				},
			},
			then: processThen{
				expectedFiles: []string{},
				expectedError: testutils.RequireErrorWithCode(specter.FileArtifactProcessorCleanUpFailedErrorCode),
			},
		},
		{
			name: "GIVEN registry entry with a non Recreate writeMode  " +
				"WHEN processing cleanup" +
				"THEN the file should not be cleaned",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{
					Files: map[string][]byte{
						"/path/to/file1": []byte("file write once"),
					},
				},
				registryEntries: []specter.ArtifactRegistryEntry{
					{
						ArtifactID: "/path/to/file1",
						Metadata: map[string]interface{}{
							"path":      "/path/to/file1",
							"writeMode": string(specter.WriteOnceMode),
						},
					},
				},
			},
			when: processWhen{
				artifacts: nil,
			},
			then: processThen{
				expectedFiles: []string{
					"/path/to/file1", // should still exist
				},
				expectedError: require.NoError,
			},
		},
		{
			name: "GIVEN registry entry with a Recreate writeMode  " +
				"WHEN processing cleanup" +
				"THEN the file should be removed",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{
					Files: map[string][]byte{
						"/path/to/file1": []byte("file to clean"),
					},
				},
				registryEntries: []specter.ArtifactRegistryEntry{
					{
						ArtifactID: "/path/to/file1",
						Metadata: map[string]interface{}{
							"path":      "/path/to/file1",
							"writeMode": string(specter.RecreateMode),
						},
					},
				},
			},
			when: processWhen{
				artifacts: nil,
			},
			then: processThen{
				expectedFiles: []string{}, // file no longer exists
				expectedError: require.NoError,
			},
		},

		// PROCESS
		{
			name: "WHEN valid file artifacts THEN file should be created successfully",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{},
			},
			when: processWhen{
				artifacts: []specter.Artifact{
					// Valid file artifacts
					&specter.FileArtifact{Path: "/path/to/file1", FileMode: os.ModePerm, WriteMode: specter.RecreateMode},
					&specter.FileArtifact{Path: "/path/to/file2", FileMode: os.ModePerm, WriteMode: specter.RecreateMode},
					specter.NewDirectoryArtifact("/path/to/file3", os.ModePerm, specter.RecreateMode),
				},
			},
			then: processThen{
				expectedFiles: []string{"/path/to/file1", "/path/to/file2", "/path/to/file3"},
				expectedError: require.NoError,
			},
		},
		{
			name: "WHEN some artifacts are not FileArtifact " +
				"THEN non FileArtifacts should be skipped and no error should be returned",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{},
			},
			when: processWhen{
				artifacts: []specter.Artifact{
					&specter.FileArtifact{Path: "/path/to/file1", FileMode: os.ModePerm},
					testutils.ArtifactStub{}, // this should be skipped.
				},
			},
			then: processThen{
				expectedFiles: []string{"/path/to/file1"},
				expectedError: require.NoError,
			},
		},
		{
			name: "GIVEN file system fails writing files THEN an error should be returned",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{
					WriteFileErr: assert.AnError,
				},
			},
			when: processWhen{
				artifacts: []specter.Artifact{
					&specter.FileArtifact{Path: "/path/to/file1", FileMode: os.ModePerm},
				},
			},
			then: processThen{
				expectedFiles: []string{},
				expectedError: testutils.RequireErrorWithCode(specter.FileArtifactProcessorProcessingFailedErrorCode),
			},
		},
		{
			name: "GIVEN file already exists WHEN write mode is Once THEN do not write file",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{
					WriteFileErr: assert.AnError, // Returning an error to make the test fail if it tries to write it.
					Files: map[string][]byte{
						"/path/to/file1": []byte("file content"), // already exists
					},
				},
			},
			when: processWhen{
				artifacts: []specter.Artifact{
					&specter.FileArtifact{Path: "/path/to/file1", FileMode: os.ModePerm, WriteMode: specter.WriteOnceMode},
				},
			},
			then: processThen{
				expectedFiles: []string{"/path/to/file1"},
				expectedError: require.NoError,
			},
		},
		{
			name: "WHEN artifact without path THEN return error",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{},
			},
			when: processWhen{
				artifacts: []specter.Artifact{
					&specter.FileArtifact{Path: "", FileMode: os.ModePerm, WriteMode: specter.WriteOnceMode},
				},
			},
			then: processThen{
				expectedError: testutils.RequireErrorWithCode(specter.FileArtifactProcessorProcessingFailedErrorCode),
			},
		},
		{
			name: "WHEN artifact without writeMode THEN no error should be returned as WRITE_ONCE will be used.",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{},
			},
			when: processWhen{
				artifacts: []specter.Artifact{
					&specter.FileArtifact{Path: "/path/to/file", FileMode: os.ModePerm, WriteMode: ""},
				},
			},
			then: processThen{
				expectedFiles: []string{
					"/path/to/file",
				},
				expectedError: require.NoError,
			},
		},
		// FILE SYSTEM FAILURES
		{
			name: "GIVEN file system cant resolve paths THEN return error",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{
					AbsErr: assert.AnError,
				},
			},
			when: processWhen{
				artifacts: []specter.Artifact{
					&specter.FileArtifact{Path: "/some/path", FileMode: os.ModePerm, WriteMode: specter.WriteOnceMode},
				},
			},
			then: processThen{
				expectedError: testutils.RequireErrorWithCode(specter.FileArtifactProcessorProcessingFailedErrorCode),
			},
		},
		{
			name: "GIVEN file system cant stat paths WHEN artifact without path THEN an error should be returned",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{
					StatErr: assert.AnError,
				},
			},
			when: processWhen{
				artifacts: []specter.Artifact{
					&specter.FileArtifact{Path: "/some/path", FileMode: 0755, WriteMode: specter.WriteOnceMode},
				},
			},
			then: processThen{
				expectedError: testutils.RequireErrorWithCode(specter.FileArtifactProcessorProcessingFailedErrorCode),
			},
		},
		{
			name: "GIVEN file system cant write files THEN an error should be returned",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{
					WriteFileErr: assert.AnError,
				},
			},
			when: processWhen{
				artifacts: []specter.Artifact{
					&specter.FileArtifact{Path: "/some/path", FileMode: 0755, WriteMode: specter.RecreateMode},
				},
			},
			then: processThen{
				expectedError: testutils.RequireErrorWithCode(specter.FileArtifactProcessorProcessingFailedErrorCode),
			},
		},
		{
			name: "GIVEN file system cant make directories THEN an error should be returned",
			given: processGiven{
				fileSystem: &testutils.MockFileSystem{
					MkdirErr: assert.AnError,
				},
			},
			when: processWhen{
				artifacts: []specter.Artifact{
					specter.NewDirectoryArtifact("/path/to/file", os.ModePerm, specter.RecreateMode),
				},
			},
			then: processThen{
				expectedError: testutils.RequireErrorWithCode(specter.FileArtifactProcessorProcessingFailedErrorCode),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := specter.FileArtifactProcessor{FileSystem: tt.given.fileSystem}

			registry := &specter.InMemoryArtifactRegistry{}
			if len(tt.given.registryEntries) != 0 {
				for _, e := range tt.given.registryEntries {
					assert.NoError(t, registry.Add(processor.Name(), e))
				}
			}

			parentCtx := context.Background()
			if tt.given.contextTimeout != 0 {
				var cancelFunc context.CancelFunc
				parentCtx, cancelFunc = context.WithTimeout(context.Background(), time.Nanosecond)
				defer cancelFunc()
			}

			ctx := specter.ArtifactProcessingContext{
				Context:          parentCtx,
				Artifacts:        tt.when.artifacts,
				Logger:           specter.NewDefaultLogger(specter.DefaultLoggerConfig{}),
				ArtifactRegistry: specter.NewProcessorArtifactRegistry(processor.Name(), registry),
			}
			err := processor.Process(ctx)

			// Test error
			if tt.then.expectedError != nil {
				tt.then.expectedError(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Test expected files
			for _, file := range tt.then.expectedFiles {
				_, err := tt.given.fileSystem.StatPath(file)
				fileExists := true
				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						fileExists = false
					}
					t.Error(err)
				}
				assert.True(t, fileExists, "file %q should have been created", file)
			}
		})
	}
}

func TestWriteFileArtifactProcessor_Process_clean(t *testing.T) {

}

func TestWriteFileArtifactProcessor_Name(t *testing.T) {
	assert.NotEqual(t, "", specter.FileArtifactProcessor{}.Name())
}

func TestNewDirectoryArtifact(t *testing.T) {
	dir := specter.NewDirectoryArtifact("/dir", os.ModePerm, specter.RecreateMode)
	require.NotNil(t, dir)
	assert.Equal(t, dir.Path, "/dir")
	assert.Equal(t, dir.FileMode, os.ModePerm|os.ModeDir)
	assert.Equal(t, dir.WriteMode, specter.RecreateMode)
	assert.Nil(t, dir.Data)
}

func TestFileArtifact_IsDir(t *testing.T) {
	f := &specter.FileArtifact{FileMode: os.ModePerm}
	assert.False(t, f.IsDir())

	f = &specter.FileArtifact{FileMode: os.ModePerm | os.ModeDir}
	assert.True(t, f.IsDir())
}
