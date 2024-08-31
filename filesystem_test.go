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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var _ FileSystem = (*mockFileSystem)(nil)

// Mock implementations to use in tests.
type mockFileInfo struct {
	os.FileInfo
	name  string
	size  int64
	mode  os.FileMode
	isDir bool
}

func (m mockFileInfo) Type() fs.FileMode {
	if m.isDir {
		return os.ModeDir
	}

	return os.ModeAppend
}

func (m mockFileInfo) Info() (fs.FileInfo, error) {
	return m, nil
}

func (m mockFileInfo) ID() string        { return m.name }
func (m mockFileInfo) Size() int64       { return m.size }
func (m mockFileInfo) Mode() os.FileMode { return m.mode }
func (m mockFileInfo) IsDir() bool       { return m.isDir }
func (m mockFileInfo) Sys() interface{}  { return nil }

type mockFileSystem struct {
	mu    sync.RWMutex
	files map[string][]byte
	dirs  map[string]bool

	absErr       error
	statErr      error
	walkDirErr   error
	readFileErr  error
	writeFileErr error
	mkdirErr     error
	rmErr        error
}

func (m *mockFileSystem) Rel(basePath, targetPath string) (string, error) {
	return strings.ReplaceAll(targetPath, basePath, "./"), nil
}

func (m *mockFileSystem) WriteFile(filePath string, data []byte, _ fs.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.writeFileErr != nil {
		return m.writeFileErr
	}

	if m.files == nil {
		m.files = map[string][]byte{}
	}

	m.files[filePath] = data

	return nil
}

func (m *mockFileSystem) Mkdir(dirPath string, _ fs.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.mkdirErr != nil {
		return m.mkdirErr
	}

	m.dirs[dirPath] = true

	return nil
}

func (m *mockFileSystem) Remove(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.rmErr != nil {
		return m.rmErr
	}

	if _, ok := m.dirs[path]; ok {
		m.dirs[path] = false
	}
	delete(m.files, path)

	return nil
}

func (m *mockFileSystem) Abs(location string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.absErr != nil {
		return "", m.absErr
	}

	absPaths := make(map[string]bool, len(m.files)+len(m.dirs))

	for k := range m.files {
		absPaths[k] = true
	}
	for k := range m.dirs {
		absPaths[k] = true
	}

	if absPath, exists := absPaths[location]; exists && absPath {
		return location, nil
	}
	return location, nil
}

func (m *mockFileSystem) StatPath(location string) (os.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.statErr != nil {
		return nil, m.statErr
	}

	if isDir, exists := m.dirs[location]; exists {
		return mockFileInfo{name: location, isDir: isDir}, nil
	}

	if data, exists := m.files[location]; exists {
		return mockFileInfo{name: location, size: int64(len(data))}, nil
	}

	return nil, os.ErrNotExist
}

func (m *mockFileSystem) WalkDir(dirPath string, f func(path string, d fs.DirEntry, err error) error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.walkDirErr != nil {
		return m.walkDirErr
	}

	for path, isDir := range m.dirs {
		if strings.HasPrefix(path, dirPath) {
			err := f(path, mockFileInfo{name: path, isDir: isDir}, nil)
			if err != nil {
				return err
			}
		}
	}

	for path := range m.files {
		if strings.HasPrefix(path, dirPath) {
			err := f(path, mockFileInfo{name: path, isDir: false}, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *mockFileSystem) ReadFile(filePath string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.readFileErr != nil {
		return nil, m.readFileErr
	}

	if data, exists := m.files[filePath]; exists {
		return data, nil
	}

	return nil, os.ErrNotExist
}

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
