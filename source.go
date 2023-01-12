package specter

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type SourceFormat string

// Source represents the source code that was used to load a given specification.
type Source struct {
	// Location of the source, this can be a local file or a remote file.
	Location string

	// RAW content of the source
	Data []byte

	// Detected format of the source.
	Format SourceFormat
}

// SourceLoader are services responsible for loading sources.
type SourceLoader interface {
	Load(location string) ([]Source, error)
	Supports(location string) bool
}

// LocalFileSourceLoader is an implementation of a SourceLoader that loads files from the local system.
type LocalFileSourceLoader struct{}

func NewLocalFileSourceLoader() LocalFileSourceLoader {
	return LocalFileSourceLoader{}
}

func (l LocalFileSourceLoader) Supports(target string) bool {
	// Get absolute path.
	location, err := filepath.Abs(target)
	if err != nil {
		return false
	}

	// Make sure file exists.
	if _, err := os.Stat(location); os.IsNotExist(err) {
		return false
	}

	return true
}

func (l LocalFileSourceLoader) Load(location string) ([]Source, error) {
	// Get absolute path.
	var err error
	location, err = filepath.Abs(location)
	if err != nil {
		return nil, errors.WrapWithMessage(err, UnsupportedSpecificationLoaderCode, fmt.Sprintf("failed loading file %s", location))
	}

	// Make sure file exists.
	stat, err := os.Stat(location)
	if os.IsNotExist(err) {
		return nil, errors.WrapWithMessage(err, UnsupportedSpecificationLoaderCode, fmt.Sprintf("failed loading file %s", location))
	}

	if stat.IsDir() {
		return l.loadDirectory(location)
	}

	return l.loadFile(location)
}

func (l LocalFileSourceLoader) loadDirectory(dirPath string) ([]Source, error) {
	var sources []Source

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		srcs, err := l.loadFile(path)
		if err != nil {
			return err
		}

		sources = append(sources, srcs...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return sources, nil
}

func (l LocalFileSourceLoader) loadFile(filePath string) ([]Source, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("failed loading file %s", filePath))
	}

	// Get format
	exts := strings.Split(path.Ext(filePath), ".")

	if len(exts) == 0 {
		return nil, errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("failed loading file %s as it has no extension", filePath))
	}
	ext := exts[len(exts)-1]
	format := SourceFormat(strings.Replace(ext, ".", "", 1))

	return []Source{
		{
			Location: filePath,
			Data:     bytes,
			Format:   format,
		},
	}, nil
}
