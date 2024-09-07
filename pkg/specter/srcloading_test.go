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
	"github.com/morebec/go-errors/errors"
	"github.com/morebec/specter/pkg/specter"
	"github.com/morebec/specter/pkg/testutils"
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
			given: "./srcloading_test.go",
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
			l := specter.NewLocalFileSourceLoader()
			require.Equal(t, tt.then, l.Supports(tt.given))
		})
	}
}

func TestLocalFileSourceLoader_Load(t *testing.T) {
	tests := []struct {
		given        specter.FileSystem
		whenLocation string
		name         string
		expectedErr  error
		expectedSrc  []specter.Source
	}{
		{
			name:         "WHEN an empty location THEN return error",
			given:        &testutils.MockFileSystem{},
			whenLocation: "",
			expectedErr:  errors.New("cannot load an empty location"),
		},
		{
			name: "WHEN a file does not exist THEN return error",
			given: &testutils.MockFileSystem{
				Files:   map[string][]byte{},
				StatErr: os.ErrNotExist,
			},
			whenLocation: "does-not-exist",
			expectedErr:  fs.ErrNotExist,
		},
		{
			name: "Given a directory with files WHEN we load the directory THEN return its files",
			given: &testutils.MockFileSystem{
				Files: map[string][]byte{
					"/fake/dir/file1.txt": []byte("file1 content"),
					"/fake/dir/file2.txt": []byte("file2 content"),
				},
				Dirs: map[string]bool{
					"/fake/dir": true,
				},
			},
			whenLocation: "/fake/dir",
			expectedSrc: []specter.Source{
				{Location: "/fake/dir/file1.txt", Data: []byte("file1 content"), Format: "txt"},
				{Location: "/fake/dir/file2.txt", Data: []byte("file2 content"), Format: "txt"},
			},
		},
		{
			name: "Given a directory with corrupted files WHEN we load the directory THEN return error",
			given: &testutils.MockFileSystem{
				Files: map[string][]byte{
					"/fake/dir/file1.txt": []byte("file1 content"),
					"/fake/dir/file2.txt": []byte("file2 content"),
				},
				Dirs: map[string]bool{
					"/fake/dir": true,
				},
				ReadFileErr: assert.AnError,
			},
			whenLocation: "/fake/dir",
			expectedSrc: []specter.Source{
				{Location: "/fake/dir/file1.txt", Data: []byte("file1 content"), Format: "txt"},
				{Location: "/fake/dir/file2.txt", Data: []byte("file2 content"), Format: "txt"},
			},
			expectedErr: assert.AnError,
		},
		{
			name: "Given cannot stat file WHEN we load the file THEN return err",
			given: &testutils.MockFileSystem{
				Files: map[string][]byte{
					"/fake/dir/file1.txt": []byte("file1 content"),
				},
				StatErr: assert.AnError,
			},
			whenLocation: "/fake/dir/file1.txt",
			expectedErr:  assert.AnError,
		},
		{

			name: "Given a file WHEN we load the file path THEN return the file",
			given: &testutils.MockFileSystem{
				Files: map[string][]byte{
					"/fake/file.txt": []byte("file content"),
				},
			},
			whenLocation: "/fake/file.txt",
			expectedSrc: []specter.Source{
				{Location: "/fake/file.txt", Data: []byte("file content"), Format: "txt"},
			},
		},
		{
			name: "GIVEN a file with no extension WHEN this file location THEN return a source without format",
			given: &testutils.MockFileSystem{
				Files: map[string][]byte{
					"file1": []byte("file1 content"),
				},
			},
			whenLocation: "file1",
			expectedSrc: []specter.Source{
				{Location: "file1", Data: []byte("file1 content"), Format: ""},
			},
		},
		{
			name: "GIVEN a corrupted file WHEN this file location THEN return error",
			given: &testutils.MockFileSystem{
				Files: map[string][]byte{
					"file1": []byte("file1 content"),
				},
				ReadFileErr: assert.AnError,
			},
			whenLocation: "file1",
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := specter.NewFileSystemSourceLoader(tt.given)
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

func TestFunctionalSourceLoader_Supports(t *testing.T) {
	t.Run("Supports is called", func(t *testing.T) {
		called := false
		l := specter.FunctionalSourceLoader{
			SupportsFunc: func(location string) bool {
				called = true
				return true
			},
		}
		got := l.Supports("")
		require.True(t, called)
		require.True(t, got)
	})

	t.Run("Supports is called", func(t *testing.T) {
		called := false
		l := specter.FunctionalSourceLoader{
			LoadFunc: func(location string) ([]specter.Source, error) {
				called = true
				return nil, nil
			},
		}
		load, err := l.Load("")

		require.True(t, called)
		require.NoError(t, err)
		require.Nil(t, load)
	})
}
