package specter

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"os"
)

type DirectoryArtifact struct {
	Path string
	Mode os.FileMode
}

type WriteDirectoryArtifactsProcessorConfig struct {
	// Add any configuration specific to directory processing if needed
	UseRegistry bool
}

type WriteDirectoryArtifactsProcessor struct {
	config WriteDirectoryArtifactsProcessorConfig
}

func NewWriteDirectoryArtifactsProcessor(conf WriteDirectoryArtifactsProcessorConfig) *WriteDirectoryArtifactsProcessor {
	return &WriteDirectoryArtifactsProcessor{
		config: conf,
	}
}

func (p WriteDirectoryArtifactsProcessor) Name() string {
	return "directory_artifacts_processor"
}

func (p WriteDirectoryArtifactsProcessor) Process(ctx ArtifactProcessingContext) error {
	ctx.Logger.Info("Creating artifact directories ...")

	var directories []DirectoryArtifact
	for _, o := range ctx.Artifacts {
		dir, ok := o.Value.(DirectoryArtifact)
		if !ok {
			continue
		}
		directories = append(directories, dir)
	}

	for _, dir := range directories {
		ctx.Logger.Info(fmt.Sprintf("Creating directory %s ...", dir.Path))
		err := os.MkdirAll(dir.Path, dir.Mode)
		if err != nil {
			ctx.Logger.Error(fmt.Sprintf("failed creating directory at %s", dir.Path))
			return errors.Wrap(err, "failed creating artifact directories")
		}
	}

	ctx.Logger.Success("Artifact directories created successfully.")
	return nil
}
