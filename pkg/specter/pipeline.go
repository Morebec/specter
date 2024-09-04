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

// PreviewMode will cause a Pipeline instance to run until the processing step only, no artifact will be processed.
const PreviewMode RunMode = "preview"

// RunThrough will cause a Pipeline instance to be run fully.
const RunThrough RunMode = "run-through"

const defaultRunMode = PreviewMode

const SourceLoadingFailedErrorCode = "specter.source_loading_failed"
const UnitLoadingFailedErrorCode = "specter.unit_loading_failed"
const UnitProcessingFailedErrorCode = "specter.unit_processing_failed"
const ArtifactProcessingFailedErrorCode = "specter.artifact_processing_failed"

// Pipeline is the service responsible to run a specter pipeline.
type Pipeline struct {
	SourceLoaders      []SourceLoader
	Loaders            []UnitLoader
	Processors         []UnitProcessor
	ArtifactProcessors []ArtifactProcessor
	ArtifactRegistry   ArtifactRegistry
	Logger             Logger
	TimeProvider       TimeProvider
}

type PipelineResult struct {
	StartedAt time.Time
	EndedAt   time.Time

	SourceLocations []string
	Sources         []Source
	Units           []Unit
	Artifacts       []Artifact
	RunMode         RunMode
}

func (r PipelineResult) ExecutionTime() time.Duration {
	return r.EndedAt.Sub(r.StartedAt)
}

// Run the pipeline from start to finish.
func (p Pipeline) Run(ctx context.Context, sourceLocations []string, runMode RunMode) (result PipelineResult, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if runMode == "" {
		p.Logger.Warning(fmt.Sprintf("No run mode provided, defaulting to %q", defaultRunMode))
		runMode = defaultRunMode
	}

	result = PipelineResult{
		StartedAt:       p.TimeProvider(),
		SourceLocations: sourceLocations,
		RunMode:         runMode,
	}

	defer func() {
		result.EndedAt = p.TimeProvider()
		p.logResult(result)
	}()

	// Load sources
	result.Sources, err = p.loadSources(ctx, sourceLocations)
	if err != nil {
		e := errors.WrapWithMessage(err, SourceLoadingFailedErrorCode, "failed loading sources")
		p.Logger.Error(e.Error())
		return result, e
	}

	// Load Units
	result.Units, err = p.loadUnits(ctx, result.Sources)
	if err != nil {
		e := errors.WrapWithMessage(err, UnitLoadingFailedErrorCode, "failed loading units")
		p.Logger.Error(e.Error())
		return result, e
	}

	// Process Units
	result.Artifacts, err = p.processUnits(ctx, result.Units)
	if err != nil {
		e := errors.WrapWithMessage(err, UnitProcessingFailedErrorCode, "failed processing units")
		p.Logger.Error(e.Error())
		return result, e
	}

	// stop here if preview
	if result.RunMode == PreviewMode {
		return result, nil
	}

	// Process Artifact
	if err = p.processArtifacts(ctx, result.Units, result.Artifacts); err != nil {
		e := errors.WrapWithMessage(err, ArtifactProcessingFailedErrorCode, "failed processing artifacts")
		p.Logger.Error(e.Error())
		return result, e
	}

	p.Logger.Success("\nProcessing completed successfully.")
	return result, nil
}

func (p Pipeline) logResult(run PipelineResult) {
	p.Logger.Info(fmt.Sprintf("\nRun Mode: %s", run.RunMode))
	p.Logger.Info(fmt.Sprintf("Started At: %s", run.StartedAt))
	p.Logger.Info(fmt.Sprintf("Ended at: %s", run.EndedAt))
	p.Logger.Info(fmt.Sprintf("Run time: %s", run.ExecutionTime()))
	p.Logger.Info(fmt.Sprintf("Number of source locations: %d", len(run.SourceLocations)))
	p.Logger.Info(fmt.Sprintf("Number of sources: %d", len(run.Sources)))
	p.Logger.Info(fmt.Sprintf("Number of units: %d", len(run.Units)))
	p.Logger.Info(fmt.Sprintf("Number of artifacts: %d", len(run.Artifacts)))
}

// loadSources only performs the Load sources step.
func (p Pipeline) loadSources(ctx context.Context, sourceLocations []string) ([]Source, error) {
	var sources []Source
	errs := errors.NewGroup(errors.InternalErrorCode)

	p.Logger.Info(fmt.Sprintf("\nLoading sources from (%d) locations:", len(sourceLocations)))
	for _, sl := range sourceLocations {
		p.Logger.Info(fmt.Sprintf("-> %q", sl))
	}

	for _, sl := range sourceLocations {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		loaded := false
		for _, l := range p.SourceLoaders {
			if l.Supports(sl) {
				loadedSources, err := l.Load(sl)
				if err != nil {
					p.Logger.Error(err.Error())
					errs = errs.Append(err)
					continue
				}
				sources = append(sources, loadedSources...)
				loaded = true
			}
		}
		if !loaded {
			p.Logger.Warning(fmt.Sprintf("source location %q was not loaded.", sl))
		}
	}

	return sources, errors.GroupOrNil(errs)
}

// loadUnits performs the loading of Units.
func (p Pipeline) loadUnits(ctx context.Context, sources []Source) ([]Unit, error) {
	p.Logger.Info("\nLoading units ...")

	// Load units
	var units []Unit
	var sourcesNotLoaded []Source
	errs := errors.NewGroup(errors.InternalErrorCode)

	for _, src := range sources {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		wasLoaded := false
		for _, l := range p.Loaders {
			if !l.SupportsSource(src) {
				continue
			}

			loadedUnits, err := l.Load(src)
			if err != nil {
				p.Logger.Error(err.Error())
				errs = errs.Append(err)
				continue
			}

			units = append(units, loadedUnits...)
			wasLoaded = true
		}

		if !wasLoaded {
			sourcesNotLoaded = append(sourcesNotLoaded, src)
		}
	}

	if len(sourcesNotLoaded) > 0 {
		for _, src := range sourcesNotLoaded {
			p.Logger.Warning(fmt.Sprintf("%q could not be loaded.", src.Location))
		}

		p.Logger.Warning("%d units were not loaded.")
	}

	p.Logger.Info(fmt.Sprintf("%d units loaded.", len(units)))

	return units, errors.GroupOrNil(errs)
}

// processUnits sends the units to processors.
func (p Pipeline) processUnits(ctx context.Context, units []Unit) ([]Artifact, error) {
	pctx := ProcessingContext{
		Context:   ctx,
		Units:     units,
		Artifacts: nil,
		Logger:    p.Logger,
	}

	p.Logger.Info("\nProcessing units ...")
	for _, processor := range p.Processors {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		artifacts, err := processor.Process(pctx)
		if err != nil {
			return nil, errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("processor %q failed", processor.Name()))
		}
		pctx.Artifacts = append(pctx.Artifacts, artifacts...)
	}

	p.Logger.Info(fmt.Sprintf("%d artifacts generated.", len(pctx.Artifacts)))
	for _, o := range pctx.Artifacts {
		p.Logger.Info(fmt.Sprintf("-> %s", o.ID()))
	}

	p.Logger.Success("Units processed successfully.")
	return pctx.Artifacts, nil
}

// processArtifacts sends a list of ProcessingArtifacts to the registered ArtifactProcessors.
func (p Pipeline) processArtifacts(ctx context.Context, units []Unit, artifacts []Artifact) error {
	if p.ArtifactRegistry == nil {
		p.ArtifactRegistry = &InMemoryArtifactRegistry{}
	}

	p.Logger.Info("\nProcessing artifacts ...")
	if err := p.ArtifactRegistry.Load(); err != nil {
		return fmt.Errorf("failed loading artifact registry: %w", err)
	}

	defer func() {
		if err := p.ArtifactRegistry.Save(); err != nil {
			p.Logger.Error(fmt.Errorf("failed saving artifact registry: %w", err).Error())
		}
	}()

	for _, processor := range p.ArtifactProcessors {
		if err := ctx.Err(); err != nil {
			return err
		}

		processorName := processor.Name()
		artifactCtx := ArtifactProcessingContext{
			Context:   ctx,
			Units:     units,
			Artifacts: artifacts,
			Logger:    p.Logger,
			ArtifactRegistry: ProcessorArtifactRegistry{
				processorName: processorName,
				registry:      p.ArtifactRegistry,
			},
			processorName: processorName,
		}

		err := processor.Process(artifactCtx)
		if err != nil {
			return errors.WrapWithMessage(err, errors.InternalErrorCode, fmt.Sprintf("artifact processor %q failed", processorName))
		}
	}

	p.Logger.Success("Artifacts processed successfully.")
	return nil
}
