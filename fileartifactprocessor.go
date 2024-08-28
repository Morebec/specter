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
	ctx.Logger.Info("Writing file artifacts ...")

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

	ctx.Logger.Success("Artifact files written successfully.")

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
		ctx.Logger.Trace(fmt.Sprintf("making directory %q  for %q ...", filePath, fa.ID()))
		if err := p.FileSystem.WriteFile(filePath, fa.Data, os.ModePerm); err != nil {
			return err
		}
	} else {
		ctx.Logger.Info(fmt.Sprintf("Writing file %q ...", filePath))
		ctx.Logger.Trace(fmt.Sprintf("creating directory %q  for %q ...", filePath, fa.ID()))
		if err := p.FileSystem.WriteFile(filePath, fa.Data, os.ModePerm); err != nil {
			return err
		}
	}

	if fa.WriteMode != WriteOnceMode {
		ctx.AddToRegistry(fa.ID())
	}

	return nil
}

func (p FileArtifactProcessor) cleanRegistry(ctx ArtifactProcessingContext) error {
	var wg sync.WaitGroup
	cleanFile := func(ctx ArtifactProcessingContext, o ArtifactID) {
		defer wg.Done()
		if err := p.FileSystem.Remove(string(o)); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return
			}
			panic(errors.Wrap(err, "failed cleaning artifact registry files"))
		}
		ctx.RemoveFromRegistry(o)
	}

	for _, o := range ctx.RegistryArtifacts() {
		wg.Add(1)
		go cleanFile(ctx, o)
	}
	wg.Wait()

	return nil
}
