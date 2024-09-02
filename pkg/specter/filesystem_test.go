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
	"io/fs"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestLocalFileSystem_ReadFile(t *testing.T) {
	tests := []struct {
		name            string
		given           string
		then            []byte
		thenErrContains string
	}{
		{
			name:            "GIVEN a file that does not exists THEN return error",
			given:           "/fake/dir/file1.txt",
			thenErrContains: "no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LocalFileSystem{}
			got, err := l.ReadFile(tt.given)

			if tt.thenErrContains != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.thenErrContains)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.then, got)
		})
	}
}

func TestLocalFileSystem_WalkDir(t *testing.T) {
	// Make sure the closure gets called.
	lfs := LocalFileSystem{}
	closureCalled := false
	err := lfs.WalkDir("/fake/dir", func(path string, d fs.DirEntry, err error) error {
		closureCalled = true
		return nil
	})
	require.True(t, closureCalled)
	require.NoError(t, err)
}

func TestLocalFileSystem_StatPath(t *testing.T) {
	lfs := LocalFileSystem{}
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	stat, err := lfs.StatPath(filename)
	require.NoError(t, err)
	assert.NotNil(t, stat)
}

func TestLocalFileSystem_Mkdir(t *testing.T) {
	lfs := LocalFileSystem{}
	dirPath := path.Join(os.TempDir(), "specter", "TestLocalFileSystem_Mkdir")
	err := lfs.Mkdir(dirPath, os.ModePerm)
	defer func(lfs LocalFileSystem, path string) {
		err = lfs.Remove(dirPath)
		require.NoError(t, err)
	}(lfs, dirPath)
	require.NoError(t, err)
}

func TestLocalFileSystem_Remove(t *testing.T) {
	lfs := LocalFileSystem{}
	dirPath := path.Join(os.TempDir(), "specter", "TestLocalFileSystem_Remove")
	err := lfs.Mkdir(dirPath, os.ModePerm)
	require.NoError(t, err)

	err = lfs.Remove(dirPath)
	require.NoError(t, err)
}

func TestLocalFileSystem_WriteFile(t *testing.T) {
	lfs := LocalFileSystem{}
	dirPath := path.Join(os.TempDir(), "specter", "TestLocalFileSystem_WriteFile")
	err := lfs.Mkdir(dirPath, os.ModePerm)
	defer func() {
		err := lfs.Remove(dirPath)
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	filePath := path.Join(dirPath, "TestLocalFileSystem_WriteFile.txt")
	err = lfs.WriteFile(filePath, []byte("hello world"), os.ModePerm)
	defer func() {
		err := lfs.Remove(filePath)
		require.NoError(t, err)
	}()
	require.NoError(t, err)
}

func TestLocalFileSystem_Rel(t *testing.T) {
	lfs := LocalFileSystem{}
	specterTestDir := path.Join(os.TempDir(), "specter")
	dirPath := path.Join(specterTestDir, "TestLocalFileSystem_Rel")
	err := lfs.Mkdir(dirPath, os.ModePerm)
	defer func() {
		err := lfs.Remove(dirPath)
		require.NoError(t, err)

		err = lfs.Remove(specterTestDir)
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	rel, err := lfs.Rel(specterTestDir, dirPath)
	require.NoError(t, err)
	assert.Equal(t, "TestLocalFileSystem_Rel", rel)
}
