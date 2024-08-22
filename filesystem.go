package specter

import (
	"io/fs"
	"os"
	"path/filepath"
)

// FileSystem is an abstraction layer over various types of file systems.
// It provides a unified interface to interact with different file system implementations,
// whether they are local, remote, or virtual.
//
// Implementations of this interface allow consumers to perform common file system operations such as
// obtaining absolute paths, retrieving file information, walking through directories, and reading files.
type FileSystem interface {
	// Abs converts a relative file path to an absolute one. The implementation may
	// differ depending on the underlying file system.
	Abs(location string) (string, error)

	// StatPath returns file information for the specified location. This typically
	// includes details like size, modification time, and whether the path is a file
	// or directory.
	StatPath(location string) (os.FileInfo, error)

	// WalkDir traverses the directory tree rooted at the specified path, calling the
	// provided function for each file or directory encountered. This allows for
	// efficient processing of large directory structures and can handle errors for
	// individual files or directories.
	WalkDir(dirPath string, f func(path string, d fs.DirEntry, err error) error) error

	// ReadFile reads the contents of the specified file and returns it as a byte
	// slice. This method abstracts away the specifics of how the file is accessed,
	// making it easier to work with files across different types of file systems.
	ReadFile(filePath string) ([]byte, error)

	// WriteFile writes data to the specified file path with the given permissions.
	// If the file exists, it will be overwritten.
	WriteFile(path string, data []byte, perm fs.FileMode) error

	// Mkdir creates a new directory at the specified path along with any necessary
	// parents, and returns nil, If the directory already exists, the implementation
	// may either return an error or ignore the request, depending on the file
	// system. This method abstracts the underlying mechanism of directory creation
	// across different file systems.
	Mkdir(dirPath string, mode fs.FileMode) error

	// Remove removes the named file or (empty) directory.
	// If there is an error, it will be of type *PathError.
	Remove(path string) error
}

var _ FileSystem = LocalFileSystem{}

// LocalFileSystem is an implementation of a FileSystem that works on the local file system where this program is running.
type LocalFileSystem struct{}

func (l LocalFileSystem) Remove(path string) error {
	return os.Remove(path)
}

func (l LocalFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (l LocalFileSystem) Mkdir(dirPath string, mode fs.FileMode) error {
	return os.MkdirAll(dirPath, mode)
}

func (l LocalFileSystem) ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func (l LocalFileSystem) WalkDir(dirPath string, f func(path string, d fs.DirEntry, err error) error) error {
	return filepath.WalkDir(dirPath, f)
}

func (l LocalFileSystem) StatPath(location string) (os.FileInfo, error) {
	return os.Stat(location)
}

func (l LocalFileSystem) Abs(location string) (string, error) {
	return filepath.Abs(location)
}
