package specter

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"os"
)

type DirectoryOutput struct {
	Path string
	Mode os.FileMode
}

type WriteDirectoryOutputsProcessorConfig struct {
	// Add any configuration specific to directory processing if needed
	UseRegistry bool
}

type WriteDirectoryOutputsProcessor struct {
	config WriteDirectoryOutputsProcessorConfig
}

func NewWriteDirectoryOutputsProcessor(conf WriteDirectoryOutputsProcessorConfig) *WriteDirectoryOutputsProcessor {
	return &WriteDirectoryOutputsProcessor{
		config: conf,
	}
}

func (p WriteDirectoryOutputsProcessor) Name() string {
	return "directory_outputs_processor"
}

func (p WriteDirectoryOutputsProcessor) Process(ctx OutputProcessingContext) error {
	ctx.Logger.Info("Creating output directories ...")

	var directories []DirectoryOutput
	for _, o := range ctx.Outputs {
		dir, ok := o.Value.(DirectoryOutput)
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
			return errors.Wrap(err, "failed creating output directories")
		}
	}

	ctx.Logger.Success("Output directories created successfully.")
	return nil
}
