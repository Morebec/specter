package specter

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"os"
	"sync"
)

const WriteFileArtifactsProcessorErrorCode = "write_file_artifacts_processor_error"

type WriteFileArtifactsProcessorConfig struct {
}

// WriteFileArtifactProcessor is a processor responsible for writing Artifact referring to files.
// To perform its work this processor looks at the processing context for any FileArtifact.
type WriteFileArtifactProcessor struct {
	config WriteFileArtifactsProcessorConfig
}

func NewWriteFilesProcessor(conf WriteFileArtifactsProcessorConfig) *WriteFileArtifactProcessor {
	return &WriteFileArtifactProcessor{
		config: conf,
	}
}

func (f WriteFileArtifactProcessor) Name() string {
	return "file_artifacts_processor"
}

// FileArtifact is a data structure that can be used by a SpecificationProcessor to artifact files that can be written by tje WriteFileArtifactProcessor.
type FileArtifact struct {
	Path string
	Data []byte
	Mode os.FileMode
}

func (f WriteFileArtifactProcessor) Process(ctx ArtifactProcessingContext) error {
	ctx.Logger.Info("Writing artifact files ...")

	var files []FileArtifact
	for _, o := range ctx.Artifacts {
		fo, ok := o.Value.(FileArtifact)
		if !ok {
			continue
		}
		files = append(files, fo)
	}

	if err := f.cleanRegistry(ctx); err != nil {
		ctx.Logger.Error("failed cleaning artifact registry")
		return err
	}

	errs := errors.NewGroup(WriteFileArtifactsProcessorErrorCode)

	// Write files concurrently to speed up process.
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(ctx ArtifactProcessingContext, file FileArtifact) {
			defer wg.Done()
			ctx.Logger.Info(fmt.Sprintf("Writing file %q ...", file.Path))
			err := os.WriteFile(file.Path, file.Data, os.ModePerm)
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("failed writing artifact file at %q", file.Path))
				errs = errs.Append(err)
			}
			ctx.AddToRegistry(file.Path)
		}(ctx, file)
	}
	wg.Wait()

	ctx.Logger.Success("Artifact files written successfully.")

	return nil
}

func (f WriteFileArtifactProcessor) cleanRegistry(ctx ArtifactProcessingContext) error {
	var wg sync.WaitGroup
	for _, o := range ctx.RegistryArtifacts() {
		wg.Add(1)
		go func(ctx ArtifactProcessingContext, o string) {
			defer wg.Done()
			if err := os.Remove(o); err != nil {
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
