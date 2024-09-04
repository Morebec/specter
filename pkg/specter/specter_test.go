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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"strings"
	"sync"
)

func RequireErrorWithCode(c string) require.ErrorAssertionFunc {
	return func(t require.TestingT, err error, i ...interface{}) {
		require.Error(t, err)

		var sysError errors.SystemError
		if !errors.As(err, &sysError) {
			t.Errorf("expected a system error with code %q but got %s", c, err)
		}
		require.Equal(t, c, sysError.Code())
	}
}

var _ specter.Unit = (*UnitStub)(nil)

type UnitStub struct {
	name     specter.UnitName
	typeName specter.UnitType
	source   specter.Source
	desc     string
}

func (us *UnitStub) Name() specter.UnitName {
	return us.name
}

func (us *UnitStub) Type() specter.UnitType {
	return us.typeName
}

func (us *UnitStub) Description() string {
	return us.desc
}

func (us *UnitStub) Source() specter.Source {
	return us.source
}

func (us *UnitStub) SetSource(src specter.Source) {
	us.source = src
}

// FILE SYSTEM
var _ specter.FileSystem = (*mockFileSystem)(nil)

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

	if m.dirs == nil {
		m.dirs = map[string]bool{}
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

// ARTIFACTS

var _ specter.Artifact = ArtifactStub{}

type ArtifactStub struct {
	id specter.ArtifactID
}

func (m ArtifactStub) ID() specter.ArtifactID {
	return m.id
}

// MockArtifactRegistry is a mock implementation of ArtifactRegistry
type MockArtifactRegistry struct {
	mock.Mock
}

func (m *MockArtifactRegistry) Load() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockArtifactRegistry) Save() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockArtifactRegistry) Add(processorName string, artifactID specter.ArtifactID) {
	m.Called(processorName, artifactID)
}

func (m *MockArtifactRegistry) Remove(processorName string, artifactID specter.ArtifactID) {
	m.Called(processorName, artifactID)
}

func (m *MockArtifactRegistry) Artifacts(processorName string) []specter.ArtifactID {
	args := m.Called(processorName)
	return args.Get(0).([]specter.ArtifactID)
}
