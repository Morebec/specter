package specter

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"os"
	"time"
)

type ExecutionMode string

// LintMode will cause a Specter instance to run until the lint step only.
const LintMode ExecutionMode = "lint"

// PreviewMode will cause a Specter instance to run until the processing step only, no output will be processed.
const PreviewMode ExecutionMode = "preview"

// FullMode will cause a Specter instance to be run fully.
const FullMode ExecutionMode = "full"

// Specter is the service responsible to run a specter pipeline.
type Specter struct {
	SourceLoaders    []SourceLoader
	Loaders          []SpecificationLoader
	Processors       []SpecificationProcessor
	OutputProcessors []OutputProcessor
	OutputRegistry   OutputRegistry
	Logger           Logger
	ExecutionMode    ExecutionMode
}

type Stats struct {
	StartedAt         time.Time
	EndedAt           time.Time
	NbSourceLocations int
	NbSources         int
	NbSpecifications  int
	NbOutputs         int
}

func (s Stats) ExecutionTime() time.Duration {
	return s.EndedAt.Sub(s.StartedAt)
}

type RunResult struct {
	Sources       []Source
	Specification []Specification
	Outputs       []ProcessingOutput
	Stats         Stats
}

// Run the pipeline from start to finish.
func (s Specter) Run(sourceLocations []string) (RunResult, error) {
	var run RunResult
	var outputs []ProcessingOutput

	defer func() {
		run.Stats.EndedAt = time.Now()
		s.Logger.Info(fmt.Sprintf("\nStarted At: %s", run.Stats.StartedAt))
		s.Logger.Info(fmt.Sprintf("Ended at: %s", run.Stats.EndedAt))
		s.Logger.Info(fmt.Sprintf("Execution time: %s", run.Stats.ExecutionTime()))
		s.Logger.Info(fmt.Sprintf("Number of source locations: %d", run.Stats.NbSourceLocations))
		s.Logger.Info(fmt.Sprintf("Number of sources: %d", run.Stats.NbSources))
		s.Logger.Info(fmt.Sprintf("Number of specifications: %d", run.Stats.NbSpecifications))
		s.Logger.Info(fmt.Sprintf("Number of outputs: %d", run.Stats.NbOutputs))
	}()

	run.Stats.StartedAt = time.Now()

	// Load sources
	run.Stats.NbSourceLocations = len(sourceLocations)
	sources, err := s.LoadSources(sourceLocations)
	run.Stats.NbSources = len(sources)
	run.Sources = sources
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading sources")
		s.Logger.Error(e.Error())
		return run, e
	}

	// Load Specifications
	var specifications []Specification
	specifications, err = s.LoadSpecifications(sources)
	run.Stats.NbSpecifications = len(specifications)
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading specifications")
		s.Logger.Error(e.Error())
		return run, e
	}

	// Process Specifications
	outputs, err = s.ProcessSpecifications(specifications)
	run.Stats.NbOutputs = len(outputs)
	run.Outputs = outputs
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed processing specifications")
		s.Logger.Error(e.Error())
		return run, e
	}
	// stop here
	if s.ExecutionMode == PreviewMode {
		return run, nil
	}

	// Process Output
	if err = s.ProcessOutputs(specifications, outputs); err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed processing outputs")
		s.Logger.Error(e.Error())
		return run, e
	}

	s.Logger.Success("\nProcessing completed successfully.")
	return run, nil
}

// LoadSources only performs the Load sources step.
func (s Specter) LoadSources(sourceLocations []string) ([]Source, error) {
	var sources []Source
	errs := errors.NewGroup(errors.InternalErrorCode)

	s.Logger.Info(fmt.Sprintf("\nLoading sources from (%d) locations:", len(sourceLocations)))
	for _, sl := range sourceLocations {
		s.Logger.Info(fmt.Sprintf("-> %q", sl))
	}

	for _, sl := range sourceLocations {
		loaded := false
		for _, l := range s.SourceLoaders {
			if l.Supports(sl) {
				loadedSources, err := l.Load(sl)
				if err != nil {
					s.Logger.Error(err.Error())
					errs = errs.Append(err)
					continue
				}
				sources = append(sources, loadedSources...)
				loaded = true
			}
		}
		if !loaded {
			s.Logger.Warning(fmt.Sprintf("source location %q was not loaded.", sl))
		}
	}

	return sources, errors.GroupOrNil(errs)
}

// LoadSpecifications performs the loading of Specifications.
func (s Specter) LoadSpecifications(sources []Source) ([]Specification, error) {
	s.Logger.Info("\nLoading specifications ...")

	// Load specifications
	var specifications []Specification
	var sourcesNotLoaded []Source
	errs := errors.NewGroup(errors.InternalErrorCode)

	for _, src := range sources {
		wasLoaded := false
		for _, l := range s.Loaders {
			if !l.SupportsSource(src) {
				continue
			}

			loadedSpecs, err := l.Load(src)
			if err != nil {
				s.Logger.Error(err.Error())
				errs = errs.Append(err)
				continue
			}

			specifications = append(specifications, loadedSpecs...)
			wasLoaded = true
		}

		if !wasLoaded {
			sourcesNotLoaded = append(sourcesNotLoaded, src)
		}
	}

	if len(sourcesNotLoaded) > 0 {
		for _, src := range sourcesNotLoaded {
			s.Logger.Warning(fmt.Sprintf("%q could not be loaded.", src))
		}

		s.Logger.Warning("%d specifications were not loaded.")
	}

	s.Logger.Info(fmt.Sprintf("%d specifications loaded.", len(specifications)))

	return specifications, errors.GroupOrNil(errs)
}

// ProcessSpecifications sends the specifications to processors.
func (s Specter) ProcessSpecifications(specs []Specification) ([]ProcessingOutput, error) {
	ctx := ProcessingContext{
		Specifications: specs,
		Outputs:        nil,
		Logger:         s.Logger,
	}

	s.Logger.Info("\nProcessing specifications ...")
	for _, p := range s.Processors {
		outputs, err := p.Process(ctx)
		if err != nil {
			return nil, errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("processor %q failed", p.Name()))
		}
		ctx.Outputs = append(ctx.Outputs, outputs...)
	}

	s.Logger.Info(fmt.Sprintf("%d outputs generated.", len(ctx.Outputs)))
	for _, o := range ctx.Outputs {
		s.Logger.Info(fmt.Sprintf("-> %s", o.Name))
	}

	s.Logger.Success("Specifications processed successfully.")
	return ctx.Outputs, nil
}

// ProcessOutputs sends a list of ProcessingOutputs to the registered OutputProcessors.
func (s Specter) ProcessOutputs(specifications []Specification, outputs []ProcessingOutput) error {
	if s.OutputRegistry == nil {
		s.OutputRegistry = NoopOutputRegistry{}
	}

	ctx := OutputProcessingContext{
		Specifications: specifications,
		Outputs:        outputs,
		Logger:         s.Logger,
		outputRegistry: s.OutputRegistry,
	}

	s.Logger.Info("\nProcessing outputs ...")
	if err := s.OutputRegistry.Load(); err != nil {
		return fmt.Errorf("failed loading output registry: %w", err)
	}

	for _, p := range s.OutputProcessors {
		err := p.Process(ctx)
		if err != nil {
			return errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("output processor %q failed", p.Name()))
		}
	}

	if err := s.OutputRegistry.Save(); err != nil {
		return fmt.Errorf("failed saving output registry: %w", err)
	}

	s.Logger.Success("Outputs processed successfully.")
	return nil
}

// New allows creating a new specter instance using the provided options.
func New(opts ...Option) *Specter {
	s := &Specter{
		Logger:        NewDefaultLogger(DefaultLoggerConfig{DisableColors: true, Writer: os.Stdout}),
		ExecutionMode: FullMode,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Option represents an option to configure a specter instance.
type Option func(s *Specter)

// WithLogger configures the Logger of a Specter instance.
func WithLogger(l Logger) Option {
	return func(s *Specter) {
		s.Logger = l
	}
}

// WithSourceLoaders configures the SourceLoader of a Specter instance.
func WithSourceLoaders(loaders ...SourceLoader) Option {
	return func(s *Specter) {
		s.SourceLoaders = append(s.SourceLoaders, loaders...)
	}
}

// WithLoaders configures the SpecificationLoader of a Specter instance.
func WithLoaders(loaders ...SpecificationLoader) Option {
	return func(s *Specter) {
		s.Loaders = append(s.Loaders, loaders...)
	}
}

// WithProcessors configures the SpecProcess of a Specter instance.
func WithProcessors(processors ...SpecificationProcessor) Option {
	return func(s *Specter) {
		s.Processors = append(s.Processors, processors...)
	}
}

// WithOutputProcessors configures the OutputProcessor of a Specter instance.
func WithOutputProcessors(processors ...OutputProcessor) Option {
	return func(s *Specter) {
		s.OutputProcessors = append(s.OutputProcessors, processors...)
	}
}

// WithExecutionMode configures the ExecutionMode of a Specter instance.
func WithExecutionMode(m ExecutionMode) Option {
	return func(s *Specter) {
		s.ExecutionMode = m
	}
}
func WithOutputRegistry(r OutputRegistry) Option {
	return func(s *Specter) {
		s.OutputRegistry = r
	}
}

// Defaults

func WithDefaultLogger() Option {
	return WithLogger(NewDefaultLogger(DefaultLoggerConfig{DisableColors: false, Writer: os.Stdout}))
}
