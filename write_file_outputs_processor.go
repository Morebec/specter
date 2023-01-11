package specter

import (
	"encoding/json"
	"fmt"
	"github.com/morebec/go-errors/errors"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

const WriteFileOutputsProcessorErrorCode = "write_file_outputs_processor_error"

// OutputFileRegistry allows tracking on the file system the files that were written by the last execution
// of the WriteFileOutputsProcessor to perform cleaning operations on next executions.
type OutputFileRegistry struct {
	GeneratedAt time.Time `json:"generatedAt"`
	Files       []string  `json:"files"`
	FilePath    string
}

// NewOutputFileRegistry returns a new output file registry.
func NewOutputFileRegistry(fileName string) OutputFileRegistry {
	return OutputFileRegistry{
		GeneratedAt: time.Now(),
		Files:       nil,
		FilePath:    fileName,
	}
}

// Load loads the registry.
func (r *OutputFileRegistry) Load() error {
	bytes, err := os.ReadFile(r.FilePath)

	if err != nil {
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading output file registry")
	}
	if err := json.Unmarshal(bytes, r); err != nil {
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading output file registry")
	}

	return nil
}

// Write will write the registry file to disk.
func (r *OutputFileRegistry) Write() error {
	if r.Files == nil {
		return nil
	}
	// Generate a JSON file containing all output files for clean up later on
	js, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed generating output file registry")
	}
	if err := ioutil.WriteFile(r.FilePath, js, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed generating output file registry")
	}

	return nil
}

// Clean will delete all files listed in the registry.
func (r *OutputFileRegistry) Clean() error {
	var wg sync.WaitGroup
	for _, f := range r.Files {
		wg.Add(1)
		f := f
		go func() {
			defer wg.Done()
			if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
				return
			}
			if err := os.Remove(f); err != nil {
				panic(errors.Wrap(err, "failed cleaning output registry files"))
			}
		}()
	}
	wg.Wait()

	return nil
}

type WriteFileOutputsProcessorConfig struct {
	// Indicates if a registry file should be used to clean up generated files when running the WriteFileOutputsProcessor.
	UseRegistry bool
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

// FileOutput is a data structure that can be used by a SpecProcessor to output files that can be written by tje WriteFileOutputsProcessor.
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

	registry := NewOutputFileRegistry(".specter.json")

	errs := errors.NewGroup(WriteFileOutputsProcessorErrorCode)
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		file := file
		go func() {
			defer wg.Done()
			ctx.Logger.Info(fmt.Sprintf("Writing file %s ...", file.Path))
			err := os.WriteFile(file.Path, file.Data, os.ModePerm)
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("failed writing output file at %s", file.Path))
				errs = errs.Append(err)
			}
			registry.Files = append(registry.Files, file.Path)
		}()
	}
	wg.Wait()

	if f.config.UseRegistry {
		ctx.Logger.Trace(fmt.Sprintf("Writing output file registry to \"%s\" ...", registry.FilePath))
		if err := registry.Write(); err != nil {
			return errors.Wrap(err, "failed writing output files")
		}
		ctx.Logger.Trace("Output file registry written successfully.")

		if errs.HasErrors() {
			return errs
		}
	}

	ctx.Logger.Success("Output files written successfully.")

	return nil
}
