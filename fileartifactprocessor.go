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
	"os"
	"sync"
)

const WriteFileArtifactsProcessorErrorCode = "write_file_artifacts_processor_error"

type WriteMode string

const (
	// RecreateMode mode indicating that the artifact should be recreated on every run.
	RecreateMode WriteMode = "RECREATE"

	// WriteOnceMode is used to indicate that a file should be created only once, and not be recreated for subsequent
	// executions of the processing. This can be useful in situations where scaffolding is required.
	WriteOnceMode WriteMode = "WRITE_ONCE"
)

const DefaultWriteMode WriteMode = WriteOnceMode

// FileArtifact is a data structure that can be used by a SpecificationProcessor to generate file artifacts
// that can be written by the FileArtifactProcessor.
type FileArtifact struct {
	Path      string
	Data      []byte
	FileMode  os.FileMode
	WriteMode WriteMode
}

func NewDirectoryArtifact(path string, fileMode os.FileMode, writeMode WriteMode) *FileArtifact {
	return &FileArtifact{
		Path:      path,
		FileMode:  fileMode | os.ModeDir,
		WriteMode: writeMode,
		Data:      nil,
	}
}

func (a FileArtifact) ID() ArtifactID {
	return ArtifactID(a.Path)
}

func (a FileArtifact) IsDir() bool {
	return a.FileMode&os.ModeDir != 0
}

// FileArtifactProcessor is a processor responsible for writing Artifact referring to files.
// To perform its work this processor looks at the processing context for any FileArtifact.
type FileArtifactProcessor struct {
	FileSystem FileSystem
}

func (p FileArtifactProcessor) Name() string {
	return "file_artifacts_processor"
}

func (p FileArtifactProcessor) Process(ctx ArtifactProcessingContext) error {
	files, err := p.findFileArtifactsFromContext(ctx)
	if err != nil {
		return err
	}

	if err := p.cleanRegistry(ctx); err != nil {
		ctx.Logger.Error("failed cleaning artifact registry")
		return err
	}

	errs := errors.NewGroup(WriteFileArtifactsProcessorErrorCode)

	// Write files concurrently to speed up process.
	ctx.Logger.Info("Writing file artifacts ...")
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, file := range files {
		if err := CheckContextDone(ctx); err != nil {
			return err
		}
		wg.Add(1)
		go func(ctx ArtifactProcessingContext, file FileArtifact) {
			defer wg.Done()
			if err := p.processFileArtifact(ctx, file); err != nil {
				ctx.Logger.Error(fmt.Sprintf("failed writing artifact file at %q", file.ID()))
				mu.Lock()
				defer mu.Unlock()
				errs = errs.Append(err)
			}
		}(ctx, file)
	}
	wg.Wait()

	if errs.HasErrors() {
		return errs
	}

	ctx.Logger.Success("Files artifacts written successfully.")

	return nil
}

func (p FileArtifactProcessor) findFileArtifactsFromContext(ctx ArtifactProcessingContext) ([]FileArtifact, error) {
	var files []FileArtifact
	var errs errors.Group

	for _, a := range ctx.Artifacts {
		fa, ok := a.(FileArtifact)
		if !ok {
			continue
		}

		if fa.WriteMode == "" {
			ctx.Logger.Trace(fmt.Sprintf("File artifact %q does not have a write mode, defaulting to %q", fa.ID(), DefaultWriteMode))
			fa.WriteMode = DefaultWriteMode
		}

		if fa.Path == "" {
			errs = errs.Append(errors.NewWithMessage(
				WriteFileArtifactsProcessorErrorCode,
				fmt.Sprintf("file artifact %q does not have a path", fa.ID()),
			))
		}

		files = append(files, fa)
	}
	return files, nil
}

func (p FileArtifactProcessor) processFileArtifact(ctx ArtifactProcessingContext, fa FileArtifact) error {
	filePath, err := p.FileSystem.Abs(fa.Path)
	if err != nil {
		return err
	}

	fileExists := true
	if _, err := p.FileSystem.StatPath(filePath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		fileExists = false
	}

	if fa.WriteMode == WriteOnceMode && fileExists {
		return nil
	}

	// At this point if the file still already exists, this means that the clean step has not
	// been executed properly.

	if fa.IsDir() {
		ctx.Logger.Info(fmt.Sprintf("Creating directory %q ...", filePath))
		ctx.Logger.Trace(fmt.Sprintf("making directory %q for %q ...", filePath, fa.ID()))
		if err := p.FileSystem.WriteFile(filePath, fa.Data, os.ModePerm); err != nil {
			return err
		}
	} else {
		ctx.Logger.Info(fmt.Sprintf("Writing file %q ...", filePath))
		ctx.Logger.Trace(fmt.Sprintf("creating file %q for %q ...", filePath, fa.ID()))
		if err := p.FileSystem.WriteFile(filePath, fa.Data, os.ModePerm); err != nil {
			return err
		}
	}

	if fa.WriteMode != WriteOnceMode {
		meta := map[string]any{
			"path":      fa.Path,
			"writeMode": fa.WriteMode,
		}
		if err := ctx.ArtifactRegistry.Add(fa.ID(), meta); err != nil {
			return err
		}
	}

	return nil
}

func (p FileArtifactProcessor) cleanRegistry(ctx ArtifactProcessingContext) error {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errs errors.Group

	ctx.Logger.Info("Cleaning file artifacts ...")
	entries, err := ctx.ArtifactRegistry.FindAll()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Metadata == nil {
			ctx.Logger.Trace(fmt.Sprintf("invalid registry entry %q: no metadata", entry.ArtifactID))
			continue
		}

		writeMode, ok := entry.Metadata["writeMode"].(WriteMode)
		if !ok {
			ctx.Logger.Trace(fmt.Sprintf("invalid registry entry %q: no write mode", entry.ArtifactID))
			continue
		}

		if writeMode != RecreateMode {
			continue
		}

		wg.Add(1)
		go func(entry ArtifactRegistryEntry) {
			defer wg.Done()
			if err := p.cleanArtifact(ctx, entry); err != nil {
				mu.Lock()
				defer mu.Unlock()
				errs = errs.Append(err)
			}
		}(entry)
	}
	wg.Wait()

	return nil
}

func (p FileArtifactProcessor) cleanArtifact(ctx ArtifactProcessingContext, entry ArtifactRegistryEntry) error {
	if entry.Metadata == nil {
		ctx.Logger.Trace(fmt.Sprintf("invalid registry entry %q: no metadata", entry.ArtifactID))
		return nil
	}
	path, ok := entry.Metadata["path"].(string)
	if !ok || path == "" {
		ctx.Logger.Trace(fmt.Sprintf("invalid registry entry %q: no path", entry.ArtifactID))
		return nil
	}

	ctx.Logger.Info(fmt.Sprintf("cleaning file artifact %q ...", entry.ArtifactID))
	if err := p.FileSystem.Remove(path); err != nil {
		return err
	}

	if err := ctx.ArtifactRegistry.Remove(entry.ArtifactID); err != nil {
		return err
	}

	return nil
}
