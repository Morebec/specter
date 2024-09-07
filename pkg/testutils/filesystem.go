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

package testutils

import (
	"io/fs"
	"os"
	"strings"
	"sync"
)

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

type MockFileSystem struct {
	mu    sync.RWMutex
	Files map[string][]byte
	Dirs  map[string]bool

	AbsErr       error
	StatErr      error
	walkDirErr   error
	ReadFileErr  error
	WriteFileErr error
	MkdirErr     error
	RmErr        error
}

func (m *MockFileSystem) Rel(basePath, targetPath string) (string, error) {
	return strings.ReplaceAll(targetPath, basePath, "./"), nil
}

func (m *MockFileSystem) WriteFile(filePath string, data []byte, _ fs.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.WriteFileErr != nil {
		return m.WriteFileErr
	}

	if m.Files == nil {
		m.Files = map[string][]byte{}
	}

	m.Files[filePath] = data

	return nil
}

func (m *MockFileSystem) Mkdir(dirPath string, _ fs.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.MkdirErr != nil {
		return m.MkdirErr
	}

	if m.Dirs == nil {
		m.Dirs = map[string]bool{}
	}

	m.Dirs[dirPath] = true

	return nil
}

func (m *MockFileSystem) Remove(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.RmErr != nil {
		return m.RmErr
	}

	if _, ok := m.Dirs[path]; ok {
		m.Dirs[path] = false
	}
	delete(m.Files, path)

	return nil
}

func (m *MockFileSystem) Abs(location string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.AbsErr != nil {
		return "", m.AbsErr
	}

	absPaths := make(map[string]bool, len(m.Files)+len(m.Dirs))

	for k := range m.Files {
		absPaths[k] = true
	}
	for k := range m.Dirs {
		absPaths[k] = true
	}

	if absPath, exists := absPaths[location]; exists && absPath {
		return location, nil
	}
	return location, nil
}

func (m *MockFileSystem) StatPath(location string) (os.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.StatErr != nil {
		return nil, m.StatErr
	}

	if isDir, exists := m.Dirs[location]; exists {
		return mockFileInfo{name: location, isDir: isDir}, nil
	}

	if data, exists := m.Files[location]; exists {
		return mockFileInfo{name: location, size: int64(len(data))}, nil
	}

	return nil, os.ErrNotExist
}

func (m *MockFileSystem) WalkDir(dirPath string, f func(path string, d fs.DirEntry, err error) error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.walkDirErr != nil {
		return m.walkDirErr
	}

	for path, isDir := range m.Dirs {
		if strings.HasPrefix(path, dirPath) {
			err := f(path, mockFileInfo{name: path, isDir: isDir}, nil)
			if err != nil {
				return err
			}
		}
	}

	for path := range m.Files {
		if strings.HasPrefix(path, dirPath) {
			err := f(path, mockFileInfo{name: path, isDir: false}, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *MockFileSystem) ReadFile(filePath string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ReadFileErr != nil {
		return nil, m.ReadFileErr
	}

	if data, exists := m.Files[filePath]; exists {
		return data, nil
	}

	return nil, os.ErrNotExist
}
