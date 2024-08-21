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

// Run the pipeline from start to finish.
func (s Specter) Run(sourceLocations []string) error {
	stats := Stats{}

	defer func() {
		stats.EndedAt = time.Now()

		s.Logger.Info(fmt.Sprintf("\nStarted At: %s", stats.StartedAt))
		s.Logger.Info(fmt.Sprintf("Ended at: %s", stats.EndedAt))
		s.Logger.Info(fmt.Sprintf("Execution time: %s", stats.ExecutionTime()))
		s.Logger.Info(fmt.Sprintf("Number of source locations: %d", stats.NbSourceLocations))
		s.Logger.Info(fmt.Sprintf("Number of sources: %d", stats.NbSources))
		s.Logger.Info(fmt.Sprintf("Number of specifications: %d", stats.NbSpecifications))
		s.Logger.Info(fmt.Sprintf("Number of outputs: %d", stats.NbOutputs))
	}()

	stats.StartedAt = time.Now()

	// Load sources
	stats.NbSourceLocations = len(sourceLocations)
	sources, err := s.LoadSources(sourceLocations)
	stats.NbSources = len(sources)
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading sources")
		s.Logger.Error(e.Error())
		return e
	}

	// Load Specifications
	var specifications []Specification
	specifications, err = s.LoadSpecifications(sources)
	stats.NbSpecifications = len(specifications)
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading specifications")
		s.Logger.Error(e.Error())
		return e
	}

	// Process Specifications
	var outputs []ProcessingOutput
	outputs, err = s.ProcessSpecifications(specifications)
	stats.NbOutputs = len(outputs)
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed processing specifications")
		s.Logger.Error(e.Error())
		return e
	}
	// stop here
	if s.ExecutionMode == PreviewMode {
		return nil
	}

	if err = s.ProcessOutputs(specifications, outputs); err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed processing outputs")
		s.Logger.Error(e.Error())
		return e
	}

	s.Logger.Success("\nProcessing completed successfully.")
	return nil
}

// LoadSources only performs the Load sources step.
func (s Specter) LoadSources(sourceLocations []string) ([]Source, error) {
	var sources []Source
	errs := errors.NewGroup(errors.InternalErrorCode)

	s.Logger.Info(fmt.Sprintf("\nLoading sources from (%d) locations:", len(sourceLocations)))
	for _, sl := range sourceLocations {
		s.Logger.Info(fmt.Sprintf("-> \"%s\"", sl))
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
			s.Logger.Warning(fmt.Sprintf("source location \"%s\" was not loaded.", sl))
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
			s.Logger.Warning(fmt.Sprintf("%s could not be loaded.", src))
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
			return nil, errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("processor \"%s\" failed", p.Name()))
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
	ctx := OutputProcessingContext{
		Specifications: specifications,
		Outputs:        outputs,
		Logger:         s.Logger,
	}

	s.Logger.Info("\nProcessing outputs ...")
	for _, p := range s.OutputProcessors {
		err := p.Process(ctx)
		if err != nil {
			return errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("output processor \"%s\" failed", p.Name()))
		}
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
