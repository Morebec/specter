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
	Linters          []SpecificationLinter
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
	return s.EndedAt.Sub(s.EndedAt)
}

// Run the pipeline from start to finish.
func (s Specter) Run(sourceLocations []string) error {
	stats := Stats{}

	defer func() {
		s.Logger.Info(fmt.Sprintf("\nStarted At: %s, ended at: %s, execution time: %s", stats.StartedAt, stats.EndedAt, stats.ExecutionTime()))
		s.Logger.Info(fmt.Sprintf("Number of source locations: %d", stats.NbSourceLocations))
		s.Logger.Info(fmt.Sprintf("Number of sources: %d", stats.NbSources))
		s.Logger.Info(fmt.Sprintf("Number of specifications: %d", stats.NbSpecifications))
		s.Logger.Info(fmt.Sprintf("Number of outputs: %d", stats.NbOutputs))
	}()

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

	// Resolve Dependencies
	var deps ResolvedDependencies
	deps, err = s.ResolveDependencies(specifications)
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "dependency resolution failed")
		s.Logger.Error(e.Error())
		return e
	}

	// Lint Specifications
	lr := s.LintSpecifications(deps)
	if lr.HasErrors() {
		errs := errors.NewGroup(errors.InternalErrorCode)
		for _, e := range lr.Errors().Errors {
			errs = errs.Append(e)
		}
		return errors.WrapWithMessage(errs, errors.InternalErrorCode, "linting errors encountered")
	}
	// stop here
	if s.ExecutionMode == LintMode {
		return nil
	}

	// Process Specifications
	var outputs []ProcessingOutput
	outputs, err = s.ProcessSpecifications(deps)
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

	if err = s.ProcessOutputs(deps, outputs); err != nil {
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
	errs := errors.NewGroup(errors.InternalErrorCode)

	for _, src := range sources {
		for _, l := range s.Loaders {
			// TODO Detect sources that were not loaded by any loader
			if l.SupportsSource(src) {
				loaded, err := l.Load(src)
				if err != nil {
					s.Logger.Error(err.Error())
					errs = errs.Append(err)
					continue
				}
				specifications = append(specifications, loaded...)
			}
		}
	}

	s.Logger.Info(fmt.Sprintf("%d specifications loaded.", len(specifications)))

	return specifications, errors.GroupOrNil(errs)
}

// ResolveDependencies resolves the dependencies between specifications.
func (s Specter) ResolveDependencies(specifications []Specification) (ResolvedDependencies, error) {
	s.Logger.Info("\nResolving dependencies...")
	deps, err := NewDependencyGraph(specifications...).Resolve()
	if err != nil {
		return nil, errors.WrapWithMessage(err, errors.InternalErrorCode, "failed resolving dependencies")
	}
	s.Logger.Success("Dependencies resolved successfully.")
	return deps, nil
}

// LintSpecifications runes all Linters against a list of Specifications.
func (s Specter) LintSpecifications(specifications []Specification) LinterResultSet {
	linter := CompositeSpecificationLinter(s.Linters...)
	s.Logger.Info("\nLinting specifications ...")
	lr := linter.Lint(specifications)
	if lr.HasWarnings() {
		for _, w := range lr.Warnings() {
			s.Logger.Warning(fmt.Sprintf("Warning: %s\n", w.Message))
		}
	}

	if lr.HasErrors() {
		for _, e := range lr.Errors().Errors {
			s.Logger.Error(fmt.Sprintf("Error: %s\n", e.Error()))
		}
	}

	if !lr.HasWarnings() && !lr.HasErrors() {
		s.Logger.Success("Specifications linted successfully.")
	}

	return lr
}

// ProcessSpecifications sends the specifications to processors.
func (s Specter) ProcessSpecifications(specifications ResolvedDependencies) ([]ProcessingOutput, error) {
	ctx := ProcessingContext{
		DependencyGraph: specifications,
		Outputs:         nil,
		Logger:          s.Logger,
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
func (s Specter) ProcessOutputs(specifications ResolvedDependencies, outputs []ProcessingOutput) error {
	ctx := OutputProcessingContext{
		DependencyGraph: specifications,
		Outputs:         outputs,
		Logger:          s.Logger,
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
		Logger:        NewColoredOutputLogger(ColoredOutputLoggerConfig{EnableColors: true, Writer: os.Stdout}),
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

// WithLinters configures the SpecificationLinter of a Specter instance.
func WithLinters(linters ...SpecificationLinter) Option {
	return func(s *Specter) {
		s.Linters = append(s.Linters, linters...)
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
