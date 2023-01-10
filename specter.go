package specter

import (
	"fmt"
	"github.com/morebec/errors-go/errors"
	"go.uber.org/zap"
)

// Specter is the service responsible to run a specter pipeline.
type Specter struct {
	SourceLoaders []SourceLoader
	SpecLoaders   []SpecLoader
	Processors    []SpecProcessor
	Linters       []SpecLinter
	Logger        zap.Logger
}

// Run the pipeline from start to finish.
func (s Specter) Run(sourceLocations []string) error {
	// Load sources

	fmt.Printf("Loading sources from (%d) locations:\n", len(sourceLocations))
	for _, sl := range sourceLocations {
		fmt.Printf("  > %s\n", sl)
	}
	sources, err := s.LoadSources(sourceLocations)
	if err != nil {
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading sources")
	}
	fmt.Println()

	// Load Specs
	fmt.Println("Loading specs ...")
	var specs []Spec
	specs, err = s.LoadSpecs(sources)
	if err != nil {
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading specs")
	}
	fmt.Printf("(%d) Specs loaded.\n", len(specs))
	fmt.Println()

	// Resolve Dependencies
	var deps ResolvedDependencies
	deps, err = s.ResolveDependencies(specs)
	if err != nil {
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "dependency resolution failed")
	}

	// Lint Specs
	lr := s.LintSpecs(deps)

	if lr.HasWarnings() {
		for _, w := range lr.Warnings() {
			fmt.Printf("Warning: %s\n", w.Message)
		}
	}

	if lr.HasErrors() {
		for _, e := range lr.Errors().Errors {
			fmt.Printf("Error: %s\n", e.Error())
		}
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "linting errors encountered")
	}

	// Process Specs
	if err = s.ProcessSpecs(deps); err != nil {
		return errors.WrapWithMessage(err, errors.InternalErrorCode, "failed processing specs")
	}

	// All good!

	return nil
}

// LoadSources only performs the Load sources step.
func (s Specter) LoadSources(sourceTargets []string) ([]Source, error) {
	var sources []Source
	errs := errors.NewGroup(errors.InternalErrorCode)

	for _, l := range s.SourceLoaders {
		// TODO Detect targets that were not loaded.
		for _, src := range sourceTargets {
			if l.Supports(src) {
				loaded, err := l.Load(src)
				if err != nil {
					errs = errs.Append(err)
					continue
				}
				sources = append(sources, loaded...)
			}
		}
	}

	return sources, errors.GroupOrNil(errs)
}

// LoadSpecs performs the loading of specs.
func (s Specter) LoadSpecs(sources []Source) ([]Spec, error) {
	// Load specs
	var specs []Spec
	errs := errors.NewGroup(errors.InternalErrorCode)

	for _, l := range s.SpecLoaders {
		// TODO Detect sources that were not loaded by any loader
		for _, src := range sources {
			if l.SupportsSource(src) {
				loaded, err := l.Load(src)
				if err != nil {
					errs = errs.Append(err)
					continue
				}
				specs = append(specs, loaded...)
			}
		}
	}

	return specs, errors.GroupOrNil(errs)
}

// ResolveDependencies resolves the dependencies between specs.
func (s Specter) ResolveDependencies(specs []Spec) (ResolvedDependencies, error) {

	deps, err := NewDependencyGraph(specs...).Resolve()
	if err != nil {
		return nil, errors.WrapWithMessage(err, errors.InternalErrorCode, "failed resolving dependencies")
	}

	return deps, nil
}

func (s Specter) LintSpecs(specs []Spec) LinterResultSet {
	linter := CompositeLinter(s.Linters...)
	return linter.Lint(specs)
}

// ProcessSpecs sends the specs to processors.
func (s Specter) ProcessSpecs(specs ResolvedDependencies) error {
	ctx := ProcessingContext{
		DependencyGraph: specs,
		Outputs:         nil,
	}

	for _, p := range s.Processors {
		outputs, err := p.Process(ctx)
		if err != nil {
			return errors.WrapWithMessage(err, errors.InternalErrorCode, "processor \"%s\" failed")
		}
		ctx.Outputs = append(ctx.Outputs, outputs...)
	}
	return nil
}

// Option represents an option to configure a specter instance.
type Option func(s *Specter)

// New allows creating a new specter instance using the provided options.
func New(opts ...Option) *Specter {
	s := &Specter{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// WithSourceLoaders allows configuring the SourceLoader of a Specter instance.
func WithSourceLoaders(loaders ...SourceLoader) Option {
	return func(s *Specter) {
		s.SourceLoaders = append(s.SourceLoaders, loaders...)
	}
}

// WithLoaders allows configuring the SpecLoader of a Specter instance.
func WithLoaders(loaders ...SpecLoader) Option {
	return func(s *Specter) {
		s.SpecLoaders = append(s.SpecLoaders, loaders...)
	}
}

// WithLinters allows configuring the SpecLinter of a Specter instance.
func WithLinters(linters ...SpecLinter) Option {
	return func(s *Specter) {
		s.Linters = append(s.Linters, linters...)
	}
}

// WithProcessors allows configuring the SpecProcess of a Specter instance.
func WithProcessors(processors ...SpecProcessor) Option {
	return func(s *Specter) {
		s.Processors = append(s.Processors, processors...)
	}
}
