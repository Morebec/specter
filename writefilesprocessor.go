package specter

import (
	"encoding/json"
	"github.com/morebec/errors-go/errors"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

// OutputFileRegistry allows tracking on the file system the files that were written by the last execution
// of the WriteFilesProcessor to perform cleaning operations on next executions.
type OutputFileRegistry struct {
	GeneratedAt time.Time `json:"generatedAt"`
	Files       []string  `json:"files"`
	filename    string
}

// NewOutputFileRegistry returns a new output file registry.
func NewOutputFileRegistry(fileName string) OutputFileRegistry {
	return OutputFileRegistry{
		GeneratedAt: time.Now(),
		Files:       nil,
		filename:    fileName,
	}
}

// Load loads the registry.
func (r *OutputFileRegistry) Load() error {
	bytes, err := os.ReadFile(r.filename)

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
	if err := ioutil.WriteFile(r.filename, js, os.ModePerm); err != nil {
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
