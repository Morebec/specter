package specter

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"os"
)

type ExecutionMode string

// LintMode will cause a Specter instance to run until the lint step only.
const LintMode ExecutionMode = "lint"

// PreviewMode will cause a Specter instance to run until the processing step only, no output will be processed.
const PreviewMode ExecutionMode = "preview"

// FullMode will cause a Specter instance to be ran fully.
const FullMode ExecutionMode = "full"

// Specter is the service responsible to run a specter pipeline.
type Specter struct {
	SourceLoaders    []SourceLoader
	SpecLoaders      []SpecLoader
	Processors       []SpecProcessor
	Linters          []SpecLinter
	OutputProcessors []OutputProcessor
	Logger           Logger
	ExecutionMode    ExecutionMode
}

// Run the pipeline from start to finish.
func (s Specter) Run(sourceLocations []string) error {
	// Load sources
	sources, err := s.LoadSources(sourceLocations)
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading sources")
		s.Logger.Error(e.Error())
		return e
	}

	// Load Specs
	var specs []Spec
	specs, err = s.LoadSpecs(sources)
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading specs")
		s.Logger.Error(e.Error())
		return e
	}

	// Resolve Dependencies
	var deps ResolvedDependencies
	deps, err = s.ResolveDependencies(specs)
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "dependency resolution failed")
		s.Logger.Error(e.Error())
		return e
	}

	// Lint Specs
	lr := s.LintSpecs(deps)
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

	// Process Specs
	var outputs []ProcessingOutput
	outputs, err = s.ProcessSpecs(deps)
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed processing specs")
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

// LoadSpecs performs the loading of specs.
func (s Specter) LoadSpecs(sources []Source) ([]Spec, error) {
	s.Logger.Info("\nLoading specs ...")

	// Load specs
	var specs []Spec
	errs := errors.NewGroup(errors.InternalErrorCode)

	for _, src := range sources {
		for _, l := range s.SpecLoaders {
			// TODO Detect sources that were not loaded by any loader
			if l.SupportsSource(src) {
				loaded, err := l.Load(src)
				if err != nil {
					s.Logger.Error(err.Error())
					errs = errs.Append(err)
					continue
				}
				specs = append(specs, loaded...)
			}
		}
	}

	s.Logger.Info(fmt.Sprintf("%d specs loaded.", len(specs)))

	return specs, errors.GroupOrNil(errs)
}

// ResolveDependencies resolves the dependencies between specs.
func (s Specter) ResolveDependencies(specs []Spec) (ResolvedDependencies, error) {
	s.Logger.Info("\nResolving dependencies...")
	deps, err := NewDependencyGraph(specs...).Resolve()
	if err != nil {
		return nil, errors.WrapWithMessage(err, errors.InternalErrorCode, "failed resolving dependencies")
	}
	s.Logger.Success("Dependencies resolved successfully.")
	return deps, nil
}

// LintSpecs runes all Linters against a list of Specs.
func (s Specter) LintSpecs(specs []Spec) LinterResultSet {
	linter := CompositeLinter(s.Linters...)
	s.Logger.Info("\nLinting specs ...")
	lr := linter.Lint(specs)
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
		s.Logger.Success("Specs linted successfully.")
	}

	return lr
}

// ProcessSpecs sends the specs to processors.
func (s Specter) ProcessSpecs(specs ResolvedDependencies) ([]ProcessingOutput, error) {
	ctx := ProcessingContext{
		DependencyGraph: specs,
		Outputs:         nil,
		Logger:          s.Logger,
	}

	s.Logger.Info("\nProcessing specs ...")
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

	s.Logger.Success("Specs processed successfully.")
	return ctx.Outputs, nil
}

// ProcessOutputs sends a list of ProcessingOutputs to the registered OutputProcessors.
func (s Specter) ProcessOutputs(specs ResolvedDependencies, outputs []ProcessingOutput) error {
	ctx := OutputProcessingContext{
		DependencyGraph: specs,
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

// WithLoaders configures the SpecLoader of a Specter instance.
func WithLoaders(loaders ...SpecLoader) Option {
	return func(s *Specter) {
		s.SpecLoaders = append(s.SpecLoaders, loaders...)
	}
}

// WithLinters configures the SpecLinter of a Specter instance.
func WithLinters(linters ...SpecLinter) Option {
	return func(s *Specter) {
		s.Linters = append(s.Linters, linters...)
	}
}

// WithProcessors configures the SpecProcess of a Specter instance.
func WithProcessors(processors ...SpecProcessor) Option {
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
