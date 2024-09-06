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

// FileArtifact is a data structure that can be used by a UnitProcessor to generate file artifacts
// that can be written by the FileArtifactProcessor.
type FileArtifact struct {
	Path      string
	Data      []byte
	FileMode  fs.FileMode
	WriteMode WriteMode
}

func NewDirectoryArtifact(path string, fileMode fs.FileMode, writeMode WriteMode) *FileArtifact {
	return &FileArtifact{
		Path:      path,
		FileMode:  fileMode | fs.ModeDir,
		WriteMode: writeMode,
		Data:      nil,
	}
}

func (a *FileArtifact) ID() ArtifactID {
	return ArtifactID(a.Path)
}

func (a *FileArtifact) IsDir() bool {
	return a.FileMode&fs.ModeDir != 0
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
		return errors.Wrap(err, FileArtifactProcessorCleanUpFailedErrorCode)
	}

	// Write files concurrently to speed up process.
	if err := p.processArtifacts(ctx, files); err != nil {
		return errors.Wrap(err, FileArtifactProcessorProcessingFailedErrorCode)
	}

	return nil
}

func (p FileArtifactProcessor) processArtifacts(ctx ArtifactProcessingContext, fileArtifacts []*FileArtifact) error {
	errs := errors.NewGroup(FileArtifactProcessorProcessingFailedErrorCode)

	// Directories need to be created before files to ensure their parent directory exists at the moment of write.
	// Separate directories and files.
	var files []*FileArtifact
	var directories []*FileArtifact
	for _, fa := range fileArtifacts {
		if fa.IsDir() {
			directories = append(directories, fa)
		} else {
			files = append(files, fa)
		}
	}

	// Create directories sequentially.
	// We delegate the responsibility of the caller to have provided the directories in the right order.
	for _, d := range directories {
		if err := p.processFileArtifact(ctx, d); err != nil {
			errs = errs.Append(err)
		}
	}

	// Process files concurrently
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file *FileArtifact) {
			defer wg.Done()
			if err := p.processFileArtifact(ctx, file); err != nil {
				mu.Lock()
				errs = errs.Append(err)
				mu.Unlock()
			}
		}(file)
	}
	wg.Wait()

	return errors.GroupOrNil(errs)
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
		return errors.NewWithMessage(FileArtifactProcessorCleanUpFailedErrorCode, "the cleanup process failed without being caught")
	}

	switch {
	case fa.IsDir():
		if err := p.FileSystem.Mkdir(filePath, fs.ModePerm); err != nil {
			return err
		}
	default:
		if err := p.FileSystem.WriteFile(filePath, fa.Data, fs.ModePerm); err != nil {
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

	entries, err := ctx.ArtifactRegistry.FindAll()
	if err != nil {
		return err
	}

	var validArtifacts []FileArtifact

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
			continue
		}

		validArtifacts = append(validArtifacts, FileArtifact{
			Path:      path,
			WriteMode: writeMode,
		})
	}

	// Proceed with cleanup
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
