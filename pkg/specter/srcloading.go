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
	"fmt"
	"github.com/morebec/go-errors/errors"
	"io/fs"
	"os"
	"path"
	"strings"
)

// SourceFormat represents the format or syntax of a source.
type SourceFormat string

// Source represents the source code that was used to load a given unit.
type Source struct {
	// Location of the source, this can be a local file or a remote file.
	Location string

	// Data is the raw content of the source.
	Data []byte

	// Format corresponds to the format that was detected for the source.
	Format SourceFormat
}

// SourceLoader are services responsible for loading sources from specific locations
// These could be a file system, or a database for example.
type SourceLoader interface {
	// Supports indicates if this loader supports a given location.
	// This function should always be called before Load.
	Supports(location string) bool

	// Load will read a source at given location and return a Source object representing the bytes read.
	// Implementations of this function can take for granted that this loader supports the given location.
	// Therefore, the Supports method should always be called on the location before it can be passed to the Load method.
	Load(location string) ([]Source, error)
}

// FileSystemSourceLoader is an implementation of a SourceLoader that loads files from a FileSystem.
type FileSystemSourceLoader struct {
	fs FileSystem
}

func (l FileSystemSourceLoader) Supports(target string) bool {
	if target == "" {
		return false
	}

	// Get absolute path.
	location, _ := l.fs.Abs(target)
	// Explicitly ignore err since filepath.Abs returns an error only if:
	// the file path is empty AND the PWD env variable returns an empty string
	// given our previous target path check, this will never happen.

	// Make sure file exists.
	if _, err := l.fs.StatPath(location); os.IsNotExist(err) {
		return false
	}

	return true
}

func (l FileSystemSourceLoader) Load(location string) ([]Source, error) {
	if location == "" {
		// This would indicate that the user forget to call the support method before calling this method.
		return nil, errors.New("cannot load an empty location")
	}

	// Get absolute path.
	var err error
	location, _ = l.fs.Abs(location)
	// Explicitly ignore err since filepath.Abs returns an error only if:
	// the file path is empty AND the PWD env variable returns an empty string
	// given our previous location path check, this will never happen.

	// Make sure file exists.
	stat, err := l.fs.StatPath(location)
	if os.IsNotExist(err) {
		return nil, errors.WrapWithMessage(err, UnsupportedSourceErrorCode, fmt.Sprintf("failed loading file %s", location))
	} else if err != nil {
		return nil, errors.WrapWithMessage(err, UnsupportedSourceErrorCode, fmt.Sprintf("failed loading file %s", location))
	}

	if stat.IsDir() {
		return l.loadDirectory(location)
	}

	return l.loadFile(location)
}

func (l FileSystemSourceLoader) loadDirectory(dirPath string) ([]Source, error) {
	var directorySources []Source

	err := l.fs.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fileSources, err := l.loadFile(path)
		if err != nil {
			return err
		}

		directorySources = append(directorySources, fileSources...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return directorySources, nil
}

func (l FileSystemSourceLoader) loadFile(filePath string) ([]Source, error) {
	bytes, err := l.fs.ReadFile(filePath)
	if err != nil {
		return nil, errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("failed loading file %s", filePath))
	}

	// Get format
	exts := strings.SplitAfter(path.Ext(filePath), ".")
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

type FunctionalSourceLoader struct {
	SupportsFunc func(location string) bool
	LoadFunc     func(location string) ([]Source, error)
}

func (f FunctionalSourceLoader) Supports(location string) bool {
	return f.SupportsFunc(location)
}

func (f FunctionalSourceLoader) Load(location string) ([]Source, error) {
	return f.LoadFunc(location)
}
