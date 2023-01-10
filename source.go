package specter

import (
	"fmt"
	"github.com/morebec/errors-go/errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type SourceFormat string

// Source represents the source code that was used to load a given spec.
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
	Load(location string) (Source, error)
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

func (l LocalFileSourceLoader) Load(location string) (Source, error) {

	// Get absolute path.
	var err error
	location, err = filepath.Abs(location)
	if err != nil {
		return Source{}, errors.WrapWithMessage(err, UnsupportedSpecLoaderCode, fmt.Sprintf("failed loading file %s", location))
	}

	// Make sure file exists.
	if _, err := os.Stat(location); os.IsNotExist(err) {
		return Source{}, errors.WrapWithMessage(err, UnsupportedSpecLoaderCode, fmt.Sprintf("failed loading file %s", location))
	}

	bytes, err := os.ReadFile(location)
	if err != nil {
		return Source{}, errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("failed loading file %s", location))
	}

	// Get format
	exts := strings.Split(path.Ext(location), ".")

	if len(exts) == 0 {
		return Source{}, errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("failed loading file %s as it has no extension", location))
	}
	ext := exts[len(exts)-1]
	format := SourceFormat(strings.Replace(ext, ".", "", 1))

	return Source{
		Location: location,
		Data:     bytes,
		Format:   format,
	}, nil
}
