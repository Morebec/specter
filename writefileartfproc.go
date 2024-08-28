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
	RecreateMode  WriteMode = "recreate"
	WriteOnceMode WriteMode = "once"
)

// FileArtifact is a data structure that can be used by a SpecificationProcessor to generate file artifacts
// that can be written by the WriteFileArtifactProcessor.
type FileArtifact struct {
	Path      string
	Data      []byte
	Mode      os.FileMode
	WriteMode WriteMode
}

func (a FileArtifact) ID() ArtifactID {
	return ArtifactID(a.Path)
}

// WriteFileArtifactProcessor is a processor responsible for writing Artifact referring to files.
// To perform its work this processor looks at the processing context for any FileArtifact.
type WriteFileArtifactProcessor struct {
	FileSystem FileSystem
}

func (p WriteFileArtifactProcessor) Name() string {
	return "file_artifacts_processor"
}

func (p WriteFileArtifactProcessor) Process(ctx ArtifactProcessingContext) error {
	ctx.Logger.Info("Writing file artifacts ...")

	var files []FileArtifact
	for _, o := range ctx.Artifacts {
		fo, ok := o.(FileArtifact)
		if !ok {
			continue
		}
		files = append(files, fo)
	}

	if err := p.cleanRegistry(ctx); err != nil {
		ctx.Logger.Error("failed cleaning artifact registry")
		return err
	}

	errs := errors.NewGroup(WriteFileArtifactsProcessorErrorCode)

	// Write files concurrently to speed up process.
	var wg sync.WaitGroup
	for _, file := range files {
		if err := CheckContextDone(ctx); err != nil {
			return err
		}
		if file.WriteMode == "" {
			file.WriteMode = WriteOnceMode
		}

		wg.Add(1)
		go func(ctx ArtifactProcessingContext, file FileArtifact) {
			defer wg.Done()

			filePath, err := p.FileSystem.Abs(file.Path)
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("failed writing artifact file at %q", filePath))
				errs = errs.Append(err)
				return
			}

			fileExists := true
			if _, err := p.FileSystem.StatPath(filePath); err != nil {
				if !os.IsNotExist(err) {
					ctx.Logger.Error(fmt.Sprintf("failed writing artifact file at %q", filePath))
					errs = errs.Append(err)
					return
				}
				fileExists = false
			}

			if file.WriteMode == WriteOnceMode && fileExists {
				ctx.Logger.Info(fmt.Sprintf("File %q already exists ... skipping", filePath))
				return
			}

			// At this point if the file still already exists, this means that the clean step has not
			// been executed properly.

			ctx.Logger.Info(fmt.Sprintf("Writing file %q ...", filePath))
			if err := p.FileSystem.WriteFile(filePath, file.Data, os.ModePerm); err != nil {
				ctx.Logger.Error(fmt.Sprintf("failed writing artifact file at %q", filePath))
				errs = errs.Append(err)
				return
			}

			if file.WriteMode != WriteOnceMode {
				ctx.AddToRegistry(file.ID())
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

func (p WriteFileArtifactProcessor) cleanRegistry(ctx ArtifactProcessingContext) error {
	var wg sync.WaitGroup
	for _, o := range ctx.RegistryArtifacts() {
		wg.Add(1)
		go func(ctx ArtifactProcessingContext, o ArtifactID) {
			defer wg.Done()
			if err := p.FileSystem.Remove(string(o)); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return
				}
				panic(errors.Wrap(err, "failed cleaning artifact registry files"))
			}
			ctx.RemoveFromRegistry(o)
		}(ctx, o)
	}
	wg.Wait()

	return nil
}
