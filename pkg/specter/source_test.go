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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"testing"
)

func TestLocalFileSourceLoader_Supports(t *testing.T) {
	tests := []struct {
		setup func()
		name  string
		given string
		then  bool
	}{
		{
			name:  "GIVEN a file that exists THEN return true",
			given: "./source_test.go",
			then:  true,
		},
		{
			name:  "GIVEN a non existent file THEN return false",
			given: "file_does_not_exist",
			then:  false,
		},
		{
			name:  "GIVEN no file path THEN return false",
			given: "",
			then:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			l := NewLocalFileSourceLoader()
			require.Equal(t, tt.then, l.Supports(tt.given))
		})
	}
}

func TestLocalFileSourceLoader_Load(t *testing.T) {
	tests := []struct {
		given        FileSystem
		whenLocation string
		name         string
		expectedErr  error
		expectedSrc  []Source
	}{
		{
			name:         "WHEN an empty location THEN return error",
			given:        &mockFileSystem{},
			whenLocation: "",
			expectedErr:  errors.New("cannot load an empty location"),
		},
		{
			name: "WHEN a file does not exist THEN return error",
			given: &mockFileSystem{
				files:   map[string][]byte{},
				statErr: os.ErrNotExist,
			},
			whenLocation: "does-not-exist",
			expectedErr:  fs.ErrNotExist,
		},
		{
			name: "Given a directory with files WHEN we load the directory THEN return its files",
			given: &mockFileSystem{
				files: map[string][]byte{
					"/fake/dir/file1.txt": []byte("file1 content"),
					"/fake/dir/file2.txt": []byte("file2 content"),
				},
				dirs: map[string]bool{
					"/fake/dir": true,
				},
			},
			whenLocation: "/fake/dir",
			expectedSrc: []Source{
				{Location: "/fake/dir/file1.txt", Data: []byte("file1 content"), Format: "txt"},
				{Location: "/fake/dir/file2.txt", Data: []byte("file2 content"), Format: "txt"},
			},
		},
		{
			name: "Given a directory with corrupted files WHEN we load the directory THEN return error",
			given: &mockFileSystem{
				files: map[string][]byte{
					"/fake/dir/file1.txt": []byte("file1 content"),
					"/fake/dir/file2.txt": []byte("file2 content"),
				},
				dirs: map[string]bool{
					"/fake/dir": true,
				},
				readFileErr: assert.AnError,
			},
			whenLocation: "/fake/dir",
			expectedSrc: []Source{
				{Location: "/fake/dir/file1.txt", Data: []byte("file1 content"), Format: "txt"},
				{Location: "/fake/dir/file2.txt", Data: []byte("file2 content"), Format: "txt"},
			},
			expectedErr: assert.AnError,
		},
		{
			name: "Given cannot stat file WHEN we load the file THEN return err",
			given: &mockFileSystem{
				files: map[string][]byte{
					"/fake/dir/file1.txt": []byte("file1 content"),
				},
				statErr: assert.AnError,
			},
			whenLocation: "/fake/dir/file1.txt",
			expectedErr:  assert.AnError,
		},
		{

			name: "Given a file WHEN we load the file path THEN return the file",
			given: &mockFileSystem{
				files: map[string][]byte{
					"/fake/file.txt": []byte("file content"),
				},
			},
			whenLocation: "/fake/file.txt",
			expectedSrc: []Source{
				{Location: "/fake/file.txt", Data: []byte("file content"), Format: "txt"},
			},
		},
		{
			name: "GIVEN a file with no extension WHEN this file location THEN return a source without format",
			given: &mockFileSystem{
				files: map[string][]byte{
					"file1": []byte("file1 content"),
				},
			},
			whenLocation: "file1",
			expectedSrc: []Source{
				{Location: "file1", Data: []byte("file1 content"), Format: ""},
			},
		},
		{
			name: "GIVEN a corrupted file WHEN this file location THEN return error",
			given: &mockFileSystem{
				files: map[string][]byte{
					"file1": []byte("file1 content"),
				},
				readFileErr: assert.AnError,
			},
			whenLocation: "file1",
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := FileSystemSourceLoader{fs: tt.given}
			src, err := loader.Load(tt.whenLocation)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedSrc, src)
		})
	}
}
