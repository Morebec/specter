package specter

import "context"

type ProcessingContext struct {
	context.Context
	Specifications SpecificationGroup
	Outputs        []ProcessingOutput
	Logger         Logger
}

// Output returns the output associated with a given processor.
func (c ProcessingContext) Output(outputName string) ProcessingOutput {
	for _, o := range c.Outputs {
		if o.Name == outputName {
			return o
		}
	}
	return ProcessingOutput{Name: outputName, Value: nil}
}

// ProcessingOutput represents an output generated by a SpecificationProcessor.
type ProcessingOutput struct {
	// Name of the Output
	Name string

	// Value of the output
	Value any
}

// SpecificationProcessor are services responsible for performing work using Specifications
// and which can possibly generate outputs.
type SpecificationProcessor interface {
	// Name returns the unique FilePath of this processor.
	Name() string

	// Process processes a group of specifications.
	Process(ctx ProcessingContext) ([]ProcessingOutput, error)
}

// OutputRegistry provides an interface for managing a registry of outputs. This
// registry tracks outputs generated during processing runs, enabling clean-up
// in subsequent runs to avoid residual artifacts and maintain a clean slate.
//
// Implementations of the OutputRegistry interface must be thread-safe to handle
// concurrent calls to TrackFile and UntrackFile methods. Multiple goroutines may
// access the registry simultaneously, so appropriate synchronization mechanisms
// should be implemented to prevent race conditions and ensure data integrity.
type OutputRegistry interface {
	// Load the registry state from persistent storage. If an error occurs, it
	// should be returned to indicate the failure of the loading operation.
	Load() error

	// Save the current state of the registry to persistent storage. If an
	// error occurs, it should be returned to indicate the failure of the saving operation.
	Save() error

	// AddOutput registers an output name under a specific processor name. This method
	// should ensure that the file path is associated with the given processor name
	// in the registry.
	AddOutput(processorName string, outputName string)

	// RemoveOutput removes a given named output registration for a specific processor name. This
	// method should ensure that the file path is disassociated from the given
	// processor name in the registry.
	RemoveOutput(processorName string, outputName string)

	// Outputs returns the outputs for a given processor.
	Outputs(processorName string) []string
}

type NoopOutputRegistry struct {
}

func (n NoopOutputRegistry) Load() error {
	return nil
}

func (n NoopOutputRegistry) Save() error {
	return nil
}

func (n NoopOutputRegistry) AddOutput(processorName string, outputName string) {}

func (n NoopOutputRegistry) RemoveOutput(processorName string, outputName string) {}

func (n NoopOutputRegistry) Outputs(processorName string) []string {
	return nil
}

type OutputProcessingContext struct {
	context.Context
	Specifications SpecificationGroup
	Outputs        []ProcessingOutput
	Logger         Logger

	outputRegistry OutputRegistry
	processorName  string
}

func (c *OutputProcessingContext) AddToRegistry(outputName string) {
	c.outputRegistry.AddOutput(c.processorName, outputName)
}

func (c *OutputProcessingContext) RemoveFromRegistry(outputName string) {
	c.outputRegistry.RemoveOutput(c.processorName, outputName)
}

func (c *OutputProcessingContext) RegistryOutputs() []string {
	return c.outputRegistry.Outputs(c.processorName)
}

// OutputProcessor are services responsible for processing outputs of SpecProcessors.
type OutputProcessor interface {
	// Process performs the processing of outputs generated by SpecificationProcessor.
	Process(ctx OutputProcessingContext) error

	// Name returns the name of this processor.
	Name() string
}

func GetContextOutput[T any](ctx ProcessingContext, name string) (v T) {
	output := ctx.Output(name)
	v, _ = output.Value.(T)
	return v
}
