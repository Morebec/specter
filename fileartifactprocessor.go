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
	"sync"
)

const FileArtifactProcessorCleanUpFailedErrorCode = "file_artifact_processor.clean_up_failed"
const FileArtifactProcessorProcessingFailedErrorCode = "file_artifact_processor.processing_failed"

type WriteMode string

const (
	// RecreateMode mode indicating that the artifact should be recreated on every run.
	RecreateMode WriteMode = "RECREATE"

	// WriteOnceMode is used to indicate that a file should be created only once, and not be recreated for subsequent
	// executions of the processing. This can be useful in situations where scaffolding is required.
	WriteOnceMode WriteMode = "WRITE_ONCE"
)

const DefaultWriteMode WriteMode = WriteOnceMode

var _ Artifact = (*FileArtifact)(nil)

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

func (a *FileArtifact) ID() ArtifactID {
	return ArtifactID(a.Path)
}

func (a *FileArtifact) IsDir() bool {
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
	files := p.findFileArtifactsFromContext(ctx)

	// Ensure context was not canceled before we start the whole process
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := p.cleanRegistry(ctx); err != nil {
		ctx.Logger.Error(fmt.Sprintf("failed cleaning artifact registry: %s", err.Error()))
		return errors.Wrap(err, FileArtifactProcessorCleanUpFailedErrorCode)
	}

	// Write files concurrently to speed up process.
	ctx.Logger.Info("Writing file artifacts ...")
	if err := p.processArtifacts(ctx, files); err != nil {
		ctx.Logger.Error(fmt.Sprintf("failed processing artifacts: %s", err.Error()))
		return errors.Wrap(err, FileArtifactProcessorProcessingFailedErrorCode)
	}

	ctx.Logger.Success("Files artifacts written successfully.")

	return nil
}

func (p FileArtifactProcessor) processArtifacts(ctx ArtifactProcessingContext, files []*FileArtifact) error {
	errs := errors.NewGroup(FileArtifactProcessorProcessingFailedErrorCode)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(ctx ArtifactProcessingContext, file *FileArtifact) {
			defer wg.Done()
			if err := p.processFileArtifact(ctx, file); err != nil {
				ctx.Logger.Error(fmt.Sprintf("failed writing artifact file at %q: %s", file.ID(), err))
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

	return nil
}

func (p FileArtifactProcessor) findFileArtifactsFromContext(ctx ArtifactProcessingContext) []*FileArtifact {
	var files []*FileArtifact

	for _, a := range ctx.Artifacts {
		fa, ok := a.(*FileArtifact)
		if !ok {
			continue
		}

		files = append(files, fa)
	}
	return files
}

func (p FileArtifactProcessor) processFileArtifact(ctx ArtifactProcessingContext, fa *FileArtifact) error {
	if fa.WriteMode == "" {
		ctx.Logger.Trace(fmt.Sprintf("File artifact %q does not have a write mode, defaulting to %q", fa.ID(), DefaultWriteMode))
		fa.WriteMode = DefaultWriteMode
	}

	if fa.Path == "" {
		return errors.NewWithMessage(
			FileArtifactProcessorProcessingFailedErrorCode,
			fmt.Sprintf("file artifact %q does not have a path", fa.ID()),
		)
	}

	filePath, err := p.FileSystem.Abs(fa.Path)
	if err != nil {
		return err
	}

	fileExists := true
	if _, err := p.FileSystem.StatPath(filePath); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		fileExists = false
	}

	if fa.WriteMode == WriteOnceMode && fileExists {
		return nil
	}

	// At this point if the file still already exists, this means that the clean step has not
	// been executed properly.
	if fileExists {
		ctx.Logger.Trace(fmt.Sprintf("the cleanup process failed without being caught: file for %q still exists", fa.ID()))
		return errors.NewWithMessage(FileArtifactProcessorCleanUpFailedErrorCode, "the cleanup process failed without being caught")
	}

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
		if err := p.addArtifactToRegistry(ctx, fa); err != nil {
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

	var validArtifacts []FileArtifact

	ctx.Logger.Trace("Validating file artifact registry entries before cleanup ...")
	// Validate registry entries before cleanup to reduce the chances of a partial cleanup
	for _, entry := range entries {
		if entry.Metadata == nil {
			return errors.NewWithMessage(
				FileArtifactProcessorCleanUpFailedErrorCode,
				fmt.Sprintf("invalid registry entry %q: no metadata", entry.ArtifactID),
			)
		}

		path, ok := entry.Metadata["path"].(string)
		if !ok || path == "" {
			// should never happen because of pre validation
			ctx.Logger.Trace(fmt.Sprintf("invalid registry entry %q: no path", entry.ArtifactID))
			return errors.NewWithMessage(
				FileArtifactProcessorCleanUpFailedErrorCode,
				fmt.Sprintf("invalid registry entry %q: no path", entry.ArtifactID),
			)
		}

		writeModeStr, ok := entry.Metadata["writeMode"].(string)
		if !ok || writeModeStr == "" {
			return errors.NewWithMessage(
				FileArtifactProcessorCleanUpFailedErrorCode,
				fmt.Sprintf("invalid registry entry %q: no write mode", entry.ArtifactID),
			)
		}

		writeMode := WriteMode(writeModeStr)
		if writeMode != RecreateMode {
			ctx.Logger.Trace(fmt.Sprintf("registry entry %q  has write mode %q, it will not be cleaned up", entry.ArtifactID, writeModeStr))
			continue
		}

		validArtifacts = append(validArtifacts, FileArtifact{
			Path:      path,
			WriteMode: writeMode,
		})
	}

	// Proceed with cleanup
	ctx.Logger.Trace("Cleaning up files of registry entries ...")
	for _, artifact := range validArtifacts {
		wg.Add(1)
		go func(artifact FileArtifact) {
			defer wg.Done()
			if err := p.cleanArtifact(ctx, artifact); err != nil {
				mu.Lock()
				defer mu.Unlock()
				errs = errs.Append(err)
			}
		}(artifact)
	}
	wg.Wait()

	return errors.GroupOrNil(errs)
}

func (p FileArtifactProcessor) cleanArtifact(ctx ArtifactProcessingContext, artifact FileArtifact) error {
	ctx.Logger.Info(fmt.Sprintf("cleaning file artifact %q ...", artifact.ID()))

	// In the case of errors here, we'd like roll back and put the registry in the same state as it was
	// to avoid having orphaned files on the file system.
	// We first remove from the registry, processThen attempt a physical removal on the file system.
	// If this removal fails, we'll add the file back to the registry so it can be retried in a next run.

	if err := ctx.ArtifactRegistry.Remove(artifact.ID()); err != nil {
		return err
	}

	if err := p.FileSystem.Remove(artifact.Path); err != nil {
		// Let's attempt a rollback of the registry entry removal
		if addBackToRegistryErr := p.addArtifactToRegistry(ctx, &artifact); addBackToRegistryErr != nil {
			// Double failure, let's return as a group.
			return errors.NewGroup(FileArtifactProcessorCleanUpFailedErrorCode, addBackToRegistryErr, err)
		}
		return err
	}

	return nil
}

func (p FileArtifactProcessor) addArtifactToRegistry(ctx ArtifactProcessingContext, fa *FileArtifact) error {
	meta := map[string]any{
		"path":      fa.Path,
		"writeMode": fa.WriteMode,
	}
	if err := ctx.ArtifactRegistry.Add(fa.ID(), meta); err != nil {
		return err
	}
	return nil
}
