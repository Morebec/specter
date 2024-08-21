package specter

import (
	"encoding/json"
	"github.com/morebec/go-errors/errors"
	"os"
	"sync"
	"time"
)

// JSONOutputRegistry allows tracking on the file system the files that were written by the last execution
// of the WriteFileOutputsProcessor to perform cleaning operations on next executions.
type JSONOutputRegistry struct {
	GeneratedAt time.Time                               `json:"generatedAt"`
	OutputMap   map[string]*JSONOutputRegistryProcessor `json:"files"`
	FilePath    string
	mu          sync.RWMutex // Mutex to protect concurrent access
}

type JSONOutputRegistryProcessor struct {
	Outputs []string `json:"files"`
}

// NewJSONOutputRegistry returns a new output file registry.
func NewJSONOutputRegistry(fileName string) *JSONOutputRegistry {
	return &JSONOutputRegistry{
		GeneratedAt: time.Now(),
		OutputMap:   nil,
		FilePath:    fileName,
	}
}

func (r *JSONOutputRegistry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	bytes, err := os.ReadFile(r.FilePath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading output file registry")
	}
	if err := json.Unmarshal(bytes, r); err != nil {
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading output file registry")
	}

	return nil
}

func (r *JSONOutputRegistry) Save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.OutputMap == nil {
		return nil
	}
	// Generate a JSON file containing all output files for clean up later on
	js, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed generating output file registry")
	}
	if err := os.WriteFile(r.FilePath, js, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed generating output file registry")
	}

	return nil
}

func (r *JSONOutputRegistry) AddOutput(processorName string, outputName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.OutputMap[processorName]; !ok {
		r.OutputMap[processorName] = &JSONOutputRegistryProcessor{}
	}
	r.OutputMap[processorName].Outputs = append(r.OutputMap[processorName].Outputs, outputName)
}

func (r *JSONOutputRegistry) RemoveOutput(processorName string, outputName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.OutputMap[processorName]; !ok {
		return
	}

	var files []string
	for _, file := range r.OutputMap[processorName].Outputs {
		if file != outputName {
			files = append(files, file)
		}
	}

	r.OutputMap[processorName].Outputs = files
}

func (r *JSONOutputRegistry) Outputs(processorName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	outputs, ok := r.OutputMap[processorName]
	if !ok {
		return nil
	}
	return outputs.Outputs
}
