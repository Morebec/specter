package specter

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"os"
	"sync"
)

const WriteFileOutputsProcessorErrorCode = "write_file_outputs_processor_error"

type WriteFileOutputsProcessorConfig struct {
}

// WriteFileOutputsProcessor is a processor responsible for writing ProcessingOutput referring to files.
// To perform its work this processor looks at the processing context for any FileOutput.
type WriteFileOutputsProcessor struct {
	config WriteFileOutputsProcessorConfig
}

func NewWriteFilesProcessor(conf WriteFileOutputsProcessorConfig) *WriteFileOutputsProcessor {
	return &WriteFileOutputsProcessor{
		config: conf,
	}
}

func (f WriteFileOutputsProcessor) Name() string {
	return "file_outputs_processor"
}

// FileOutput is a data structure that can be used by a SpecificationProcessor to output files that can be written by tje WriteFileOutputsProcessor.
type FileOutput struct {
	Path string
	Data []byte
	Mode os.FileMode
}

func (f WriteFileOutputsProcessor) Process(ctx OutputProcessingContext) error {
	ctx.Logger.Info("Writing output files ...")

	var files []FileOutput
	for _, o := range ctx.Outputs {
		fo, ok := o.Value.(FileOutput)
		if !ok {
			continue
		}
		files = append(files, fo)
	}

	if err := f.cleanRegistry(ctx); err != nil {
		ctx.Logger.Error("failed cleaning output registry")
		return err
	}

	errs := errors.NewGroup(WriteFileOutputsProcessorErrorCode)

	// Write files concurrently to speed up process.
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(ctx OutputProcessingContext, file FileOutput) {
			defer wg.Done()
			ctx.Logger.Info(fmt.Sprintf("Writing file %q ...", file.Path))
			err := os.WriteFile(file.Path, file.Data, os.ModePerm)
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("failed writing output file at %q", file.Path))
				errs = errs.Append(err)
			}
			ctx.AddToRegistry(file.Path)
		}(ctx, file)
	}
	wg.Wait()

	ctx.Logger.Success("Output files written successfully.")

	return nil
}

func (f WriteFileOutputsProcessor) cleanRegistry(ctx OutputProcessingContext) error {
	var wg sync.WaitGroup
	for _, o := range ctx.RegistryOutputs() {
		wg.Add(1)
		go func(ctx OutputProcessingContext, o string) {
			defer wg.Done()
			if err := os.Remove(o); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return
				}
				panic(errors.Wrap(err, "failed cleaning output registry files"))
			}
			ctx.RemoveFromRegistry(o)
		}(ctx, o)
	}
	wg.Wait()

	return nil
}
