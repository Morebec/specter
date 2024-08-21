package specter

import (
	"github.com/morebec/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"runtime"
	"strings"
	"testing"
)

//func TestLocalFileSourceLoader_Load(t *testing.T) {
//	existingFile := "./source_test.spec.hcl"
//	absPath, _ := filepath.Abs(existingFile)
//
//	type args struct {
//		target string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    []Source
//		wantErr bool
//	}{
//		{
//			name: "non existing file should return error",
//			args: args{
//				target: "does-not-exist",
//			},
//			want:    nil,
//			wantErr: true,
//		},
//		{
//			name: "existing file should return valid source",
//			args: args{
//				target: existingFile,
//			},
//			want: []Source{
//				{
//					Location: absPath,
//					Data:     []byte{},
//					Format:   HCLSourceFormat,
//				},
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			l := FileSystemLoader{}
//			got, err := l.Load(tt.args.target)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("Load() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

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
			loader := FileSystemLoader{fs: tt.given}
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

// Mock implementations to use in tests.
type mockFileInfo struct {
	os.FileInfo
	name    string
	size    int64
	mode    os.FileMode
	modTime int64
	isDir   bool
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

func (m mockFileInfo) Name() string      { return m.name }
func (m mockFileInfo) Size() int64       { return m.size }
func (m mockFileInfo) Mode() os.FileMode { return m.mode }
func (m mockFileInfo) IsDir() bool       { return m.isDir }
func (m mockFileInfo) Sys() interface{}  { return nil }

type mockFileSystem struct {
	files map[string][]byte
	dirs  map[string]bool

	absErr      error
	statErr     error
	walkDirErr  error
	readFileErr error
}

func (m *mockFileSystem) Abs(location string) (string, error) {
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
	return "", nil
}

func (m *mockFileSystem) StatPath(location string) (os.FileInfo, error) {
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
