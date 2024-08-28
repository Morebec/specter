package specter

import (
	"context"
	"fmt"
	"github.com/morebec/go-errors/errors"
	"os"
	"time"
)

type ExecutionMode string

// LintMode will cause a Specter instance to run until the lint step only.
const LintMode ExecutionMode = "lint"

// PreviewMode will cause a Specter instance to run until the processing step only, no artifact will be processed.
const PreviewMode ExecutionMode = "preview"

// FullMode will cause a Specter instance to be run fully.
const FullMode ExecutionMode = "full"

// Specter is the service responsible to run a specter pipeline.
type Specter struct {
	SourceLoaders      []SourceLoader
	Loaders            []SpecificationLoader
	Processors         []SpecificationProcessor
	ArtifactProcessors []ArtifactProcessor
	ArtifactRegistry   ArtifactRegistry
	Logger             Logger
	ExecutionMode      ExecutionMode
}

type Stats struct {
	StartedAt         time.Time
	EndedAt           time.Time
	NbSourceLocations int
	NbSources         int
	NbSpecifications  int
	NbArtifacts       int
}

func (s Stats) ExecutionTime() time.Duration {
	return s.EndedAt.Sub(s.StartedAt)
}

type RunResult struct {
	Sources       []Source
	Specification []Specification
	Artifacts     []Artifact
	Stats         Stats
}

// Run the pipeline from start to finish.
func (s Specter) Run(ctx context.Context, sourceLocations []string) (RunResult, error) {
	var run RunResult
	var artifacts []Artifact

	defer func() {
		run.Stats.EndedAt = time.Now()
		s.Logger.Info(fmt.Sprintf("\nStarted At: %s", run.Stats.StartedAt))
		s.Logger.Info(fmt.Sprintf("Ended at: %s", run.Stats.EndedAt))
		s.Logger.Info(fmt.Sprintf("Execution time: %s", run.Stats.ExecutionTime()))
		s.Logger.Info(fmt.Sprintf("Number of source locations: %d", run.Stats.NbSourceLocations))
		s.Logger.Info(fmt.Sprintf("Number of sources: %d", run.Stats.NbSources))
		s.Logger.Info(fmt.Sprintf("Number of specifications: %d", run.Stats.NbSpecifications))
		s.Logger.Info(fmt.Sprintf("Number of artifacts: %d", run.Stats.NbArtifacts))
	}()

	run.Stats.StartedAt = time.Now()

	// Load sources
	run.Stats.NbSourceLocations = len(sourceLocations)
	sources, err := s.LoadSources(ctx, sourceLocations)
	run.Stats.NbSources = len(sources)
	run.Sources = sources
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading sources")
		s.Logger.Error(e.Error())
		return run, e
	}

	// Load Specifications
	var specifications []Specification
	specifications, err = s.LoadSpecifications(ctx, sources)
	run.Stats.NbSpecifications = len(specifications)
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed loading specifications")
		s.Logger.Error(e.Error())
		return run, e
	}

	// Process Specifications
	artifacts, err = s.ProcessSpecifications(ctx, specifications)
	run.Stats.NbArtifacts = len(artifacts)
	run.Artifacts = artifacts
	if err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed processing specifications")
		s.Logger.Error(e.Error())
		return run, e
	}
	// stop here
	if s.ExecutionMode == PreviewMode {
		return run, nil
	}

	// Process Artifact
	if err = s.ProcessArtifacts(ctx, specifications, artifacts); err != nil {
		e := errors.WrapWithMessage(err, errors.InternalErrorCode, "failed processing artifacts")
		s.Logger.Error(e.Error())
		return run, e
	}

	s.Logger.Success("\nProcessing completed successfully.")
	return run, nil
}

// LoadSources only performs the Load sources step.
func (s Specter) LoadSources(ctx context.Context, sourceLocations []string) ([]Source, error) {
	var sources []Source
	errs := errors.NewGroup(errors.InternalErrorCode)

	s.Logger.Info(fmt.Sprintf("\nLoading sources from (%d) locations:", len(sourceLocations)))
	for _, sl := range sourceLocations {
		s.Logger.Info(fmt.Sprintf("-> %q", sl))
	}

	for _, sl := range sourceLocations {
		if err := CheckContextDone(ctx); err != nil {
			return nil, err
		}

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
func (s Specter) LoadSpecifications(ctx context.Context, sources []Source) ([]Specification, error) {
	s.Logger.Info("\nLoading specifications ...")

	// Load specifications
	var specifications []Specification
	var sourcesNotLoaded []Source
	errs := errors.NewGroup(errors.InternalErrorCode)

	for _, src := range sources {
		if err := CheckContextDone(ctx); err != nil {
			return nil, err
		}
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
func (s Specter) ProcessSpecifications(ctx context.Context, specs []Specification) ([]Artifact, error) {
	pctx := ProcessingContext{
		Context:        ctx,
		Specifications: specs,
		Artifacts:      nil,
		Logger:         s.Logger,
	}

	s.Logger.Info("\nProcessing specifications ...")
	for _, p := range s.Processors {
		if err := CheckContextDone(ctx); err != nil {
			return nil, err
		}
		artifacts, err := p.Process(pctx)
		if err != nil {
			return nil, errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("processor %q failed", p.Name()))
		}
		pctx.Artifacts = append(pctx.Artifacts, artifacts...)
	}

	s.Logger.Info(fmt.Sprintf("%d artifacts generated.", len(pctx.Artifacts)))
	for _, o := range pctx.Artifacts {
		s.Logger.Info(fmt.Sprintf("-> %s", o.ID()))
	}

	s.Logger.Success("Specifications processed successfully.")
	return pctx.Artifacts, nil
}

// ProcessArtifacts sends a list of ProcessingArtifacts to the registered ArtifactProcessors.
func (s Specter) ProcessArtifacts(ctx context.Context, specifications []Specification, artifacts []Artifact) error {
	if s.ArtifactRegistry == nil {
		s.ArtifactRegistry = NoopArtifactRegistry{}
	}

	s.Logger.Info("\nProcessing artifacts ...")
	if err := s.ArtifactRegistry.Load(); err != nil {
		return fmt.Errorf("failed loading artifact registry: %w", err)
	}

	for _, p := range s.ArtifactProcessors {
		if err := CheckContextDone(ctx); err != nil {
			return err
		}

		octx := ArtifactProcessingContext{
			Context:          ctx,
			Specifications:   specifications,
			Artifacts:        artifacts,
			Logger:           s.Logger,
			artifactRegistry: s.ArtifactRegistry,
			processorName:    p.Name(),
		}

		err := p.Process(octx)
		if err != nil {
			return errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("artifact processor %q failed", p.Name()))
		}
	}

	if err := s.ArtifactRegistry.Save(); err != nil {
		return fmt.Errorf("failed saving artifact registry: %w", err)
	}

	s.Logger.Success("Artifacts processed successfully.")
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

// WithArtifactProcessors configures the ArtifactProcessor of a Specter instance.
func WithArtifactProcessors(processors ...ArtifactProcessor) Option {
	return func(s *Specter) {
		s.ArtifactProcessors = append(s.ArtifactProcessors, processors...)
	}
}

// WithExecutionMode configures the ExecutionMode of a Specter instance.
func WithExecutionMode(m ExecutionMode) Option {
	return func(s *Specter) {
		s.ExecutionMode = m
	}
}
func WithArtifactRegistry(r ArtifactRegistry) Option {
	return func(s *Specter) {
		s.ArtifactRegistry = r
	}
}

// Defaults

func WithDefaultLogger() Option {
	return WithLogger(NewDefaultLogger(DefaultLoggerConfig{DisableColors: false, Writer: os.Stdout}))
}
