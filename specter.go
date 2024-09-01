// Copyright 2024 Morébec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package specter

import (
	"context"
	"fmt"
	"github.com/morebec/go-errors/errors"
	"time"
)

type RunMode string

// PreviewMode will cause a Specter instance to run until the processing step only, no artifact will be processed.
const PreviewMode RunMode = "preview"

// RunThrough will cause a Specter instance to be run fully.
const RunThrough RunMode = "run-through"

const defaultRunMode = PreviewMode

const SourceLoadingFailedErrorCode = "specter.source_loading_failed"
const SpecificationLoadingFailedErrorCode = "specter.specification_loading_failed"
const SpecificationProcessingFailedErrorCode = "specter.specification_processing_failed"
const ArtifactProcessingFailedErrorCode = "specter.artifact_processing_failed"

// Specter is the service responsible to run a specter pipeline.
type Specter struct {
	SourceLoaders      []SourceLoader
	Loaders            []SpecificationLoader
	Processors         []SpecificationProcessor
	ArtifactProcessors []ArtifactProcessor
	ArtifactRegistry   ArtifactRegistry
	Logger             Logger
	TimeProvider       TimeProvider
}

type RunResult struct {
	StartedAt time.Time
	EndedAt   time.Time

	SourceLocations []string
	Sources         []Source
	Specifications  []Specification
	Artifacts       []Artifact
	RunMode         RunMode
}

func (r RunResult) ExecutionTime() time.Duration {
	return r.EndedAt.Sub(r.StartedAt)
}

// Run the pipeline from start to finish.
func (s Specter) Run(ctx context.Context, sourceLocations []string, runMode RunMode) (run RunResult, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if runMode == "" {
		s.Logger.Warning(fmt.Sprintf("No run mode provided, defaulting to %q", defaultRunMode))
		runMode = defaultRunMode
	}

	run = RunResult{
		StartedAt:       s.TimeProvider(),
		SourceLocations: sourceLocations,
		RunMode:         runMode,
	}

	defer func() {
		run.EndedAt = s.TimeProvider()
		s.logRunResult(run)
	}()

	// Load sources
	run.Sources, err = s.loadSources(ctx, sourceLocations)
	if err != nil {
		e := errors.WrapWithMessage(err, SourceLoadingFailedErrorCode, "failed loading sources")
		s.Logger.Error(e.Error())
		return run, e
	}

	// Load Specifications
	run.Specifications, err = s.loadSpecifications(ctx, run.Sources)
	if err != nil {
		e := errors.WrapWithMessage(err, SpecificationLoadingFailedErrorCode, "failed loading specifications")
		s.Logger.Error(e.Error())
		return run, e
	}

	// Process Specifications
	run.Artifacts, err = s.processSpecifications(ctx, run.Specifications)
	if err != nil {
		e := errors.WrapWithMessage(err, SpecificationProcessingFailedErrorCode, "failed processing specifications")
		s.Logger.Error(e.Error())
		return run, e
	}

	// stop here if preview
	if run.RunMode == PreviewMode {
		return run, nil
	}

	// Process Artifact
	if err = s.processArtifacts(ctx, run.Specifications, run.Artifacts); err != nil {
		e := errors.WrapWithMessage(err, ArtifactProcessingFailedErrorCode, "failed processing artifacts")
		s.Logger.Error(e.Error())
		return run, e
	}

	s.Logger.Success("\nProcessing completed successfully.")
	return run, nil
}

func (s Specter) logRunResult(run RunResult) {
	s.Logger.Info(fmt.Sprintf("Run Mode: %s", run.RunMode))
	s.Logger.Info(fmt.Sprintf("\nStarted At: %s", run.StartedAt))
	s.Logger.Info(fmt.Sprintf("Ended at: %s", run.EndedAt))
	s.Logger.Info(fmt.Sprintf("Run time: %s", run.ExecutionTime()))
	s.Logger.Info(fmt.Sprintf("Number of source locations: %d", len(run.SourceLocations)))
	s.Logger.Info(fmt.Sprintf("Number of sources: %d", len(run.Sources)))
	s.Logger.Info(fmt.Sprintf("Number of specifications: %d", len(run.Specifications)))
	s.Logger.Info(fmt.Sprintf("Number of artifacts: %d", len(run.Artifacts)))
}

// loadSources only performs the Load sources step.
func (s Specter) loadSources(ctx context.Context, sourceLocations []string) ([]Source, error) {
	var sources []Source
	errs := errors.NewGroup(errors.InternalErrorCode)

	s.Logger.Info(fmt.Sprintf("\nLoading sources from (%d) locations:", len(sourceLocations)))
	for _, sl := range sourceLocations {
		s.Logger.Info(fmt.Sprintf("-> %q", sl))
	}

	for _, sl := range sourceLocations {
		if err := ctx.Err(); err != nil {
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

// loadSpecifications performs the loading of Specifications.
func (s Specter) loadSpecifications(ctx context.Context, sources []Source) ([]Specification, error) {
	s.Logger.Info("\nLoading specifications ...")

	// Load specifications
	var specifications []Specification
	var sourcesNotLoaded []Source
	errs := errors.NewGroup(errors.InternalErrorCode)

	for _, src := range sources {
		if err := ctx.Err(); err != nil {
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

// processSpecifications sends the specifications to processors.
func (s Specter) processSpecifications(ctx context.Context, specs []Specification) ([]Artifact, error) {
	pctx := ProcessingContext{
		Context:        ctx,
		Specifications: specs,
		Artifacts:      nil,
		Logger:         s.Logger,
	}

	s.Logger.Info("\nProcessing specifications ...")
	for _, p := range s.Processors {
		if err := ctx.Err(); err != nil {
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

// processArtifacts sends a list of ProcessingArtifacts to the registered ArtifactProcessors.
func (s Specter) processArtifacts(ctx context.Context, specifications []Specification, artifacts []Artifact) error {
	if s.ArtifactRegistry == nil {
		s.ArtifactRegistry = &InMemoryArtifactRegistry{}
	}

	s.Logger.Info("\nProcessing artifacts ...")
	if err := s.ArtifactRegistry.Load(); err != nil {
		return fmt.Errorf("failed loading artifact registry: %w", err)
	}

	defer func() {
		if err := s.ArtifactRegistry.Save(); err != nil {
			s.Logger.Error(fmt.Errorf("failed saving artifact registry: %w", err).Error())
		}
	}()

	for _, p := range s.ArtifactProcessors {
		if err := ctx.Err(); err != nil {
			return err
		}

		artifactCtx := ArtifactProcessingContext{
			Context:        ctx,
			Specifications: specifications,
			Artifacts:      artifacts,
			Logger:         s.Logger,
			ArtifactRegistry: ProcessorArtifactRegistry{
				processorName: p.Name(),
				registry:      s.ArtifactRegistry,
			},
			processorName: p.Name(),
		}

		err := p.Process(artifactCtx)
		if err != nil {
			return errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("artifact processor %q failed", p.Name()))
		}
	}

	s.Logger.Success("Artifacts processed successfully.")
	return nil
}
