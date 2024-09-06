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
)

// DefaultPipeline is the service responsible to run a specter DefaultPipeline.
type DefaultPipeline struct {
	TimeProvider TimeProvider

	SourceLoadingStage      SourceLoadingStage
	UnitLoadingStage        UnitLoadingStage
	UnitProcessingStage     UnitProcessingStage
	ArtifactProcessingStage ArtifactProcessingStage
}

// Run the DefaultPipeline from start to finish.
func (p DefaultPipeline) Run(ctx context.Context, sourceLocations []string, runMode RunMode) (PipelineResult, error) {
	if runMode == "" {
		runMode = RunThrough
	}

	pctx := &PipelineContext{
		Context: ctx,
		PipelineContextData: PipelineContextData{
			StartedAt:       p.TimeProvider(),
			SourceLocations: sourceLocations,
			RunMode:         runMode,
		},
	}

	err := p.run(pctx, sourceLocations, runMode)

	result := PipelineResult{
		PipelineContextData: pctx.PipelineContextData,
		EndedAt:             p.TimeProvider(),
	}

	return result, err
}

func (p DefaultPipeline) run(pctx *PipelineContext, sourceLocations []string, runMode RunMode) error {
	if err := p.runSourceLoadingStage(pctx, sourceLocations); err != nil {
		return err
	}
	if runMode == StopAfterSourceLoadingStage {
		return nil
	}

	if err := p.runUnitLoadingStage(pctx); err != nil {
		return err
	}
	if runMode == StopAfterUnitLoadingStage {
		return nil
	}

	if err := p.runUnitProcessingStage(pctx); err != nil {
		return err
	}
	if runMode == StopAfterUnitProcessingStage {
		return nil
	}

	if err := p.runArtifactProcessingStage(pctx); err != nil {
		return err
	}
	return nil
}

func (p DefaultPipeline) runSourceLoadingStage(pctx *PipelineContext, sourceLocations []string) error {
	if err := pctx.Err(); err != nil {
		return err
	}

	var err error
	if p.SourceLoadingStage != nil {
		pctx.Sources, err = p.SourceLoadingStage.Run(*pctx, sourceLocations)
		if err != nil {
			return errors.WrapWithMessage(err, SourceLoadingFailedErrorCode, "failed loading sources")
		}
	}

	return nil
}

func (p DefaultPipeline) runUnitLoadingStage(pctx *PipelineContext) error {
	if err := pctx.Err(); err != nil {
		return err
	}

	var err error
	if p.UnitLoadingStage != nil {
		pctx.Units, err = p.UnitLoadingStage.Run(*pctx, pctx.Sources)
		if err != nil {
			return errors.WrapWithMessage(err, UnitLoadingFailedErrorCode, "failed loading units")
		}
	}

	return nil
}

func (p DefaultPipeline) runUnitProcessingStage(pctx *PipelineContext) error {
	if err := pctx.Err(); err != nil {
		return err
	}

	var err error
	if p.UnitProcessingStage != nil {
		pctx.Artifacts, err = p.UnitProcessingStage.Run(*pctx, pctx.Units)
		if err != nil {
			return errors.WrapWithMessage(err, UnitProcessingFailedErrorCode, "failed processing units")
		}
	}
	return nil
}

func (p DefaultPipeline) runArtifactProcessingStage(pctx *PipelineContext) error {
	if err := pctx.Err(); err != nil {
		return err
	}

	if p.ArtifactProcessingStage != nil {
		if err := p.ArtifactProcessingStage.Run(*pctx, pctx.Artifacts); err != nil {
			return errors.WrapWithMessage(err, ArtifactProcessingFailedErrorCode, "failed processing artifacts")
		}
	}
	return nil
}

type SourceLoadingStageHooksAdapter struct{}

func (_ SourceLoadingStageHooksAdapter) Before(_ PipelineContext) error { return nil }
func (_ SourceLoadingStageHooksAdapter) After(_ PipelineContext) error  { return nil }
func (_ SourceLoadingStageHooksAdapter) BeforeSourceLocation(_ PipelineContext, _ string) error {
	return nil
}
func (_ SourceLoadingStageHooksAdapter) AfterSourceLocation(_ PipelineContext, _ string) error {
	return nil
}
func (_ SourceLoadingStageHooksAdapter) OnError(_ PipelineContext, err error) error {
	return err
}

type sourceLoadingStage struct {
	SourceLoaders []SourceLoader
	Hooks         SourceLoadingStageHooks
}

func (s sourceLoadingStage) Run(ctx PipelineContext, sourceLocations []string) ([]Source, error) {
	if s.Hooks == nil {
		s.Hooks = SourceLoadingStageHooksAdapter{}
	}

	if err := s.Hooks.Before(ctx); err != nil {
		return nil, newFailedToRunHookErr(err, "Before")
	}

	sources, err := s.run(ctx, sourceLocations)
	if err != nil {
		err = errors.WrapWithMessage(err, SourceLoadingFailedErrorCode, "failed to load sources")
		return nil, s.Hooks.OnError(ctx, err)
	}

	if err := s.Hooks.After(ctx); err != nil {
		return nil, newFailedToRunHookErr(err, "After")
	}

	return sources, nil
}

func (s sourceLoadingStage) run(ctx PipelineContext, sourceLocations []string) ([]Source, error) {
	ctx.SourceLocations = sourceLocations

	errs := errors.NewGroup(SourceLoadingFailedErrorCode)

	for _, sl := range sourceLocations {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if err := s.Hooks.BeforeSourceLocation(ctx, sl); err != nil {
			return nil, newFailedToRunHookErr(err, "BeforeSourceLocation")
		}

		sources, err := s.processSourceLocation(ctx, sl)
		if err != nil {
			errs = errs.Append(err)
			continue
		}
		ctx.Sources = append(ctx.Sources, sources...)

		if err := s.Hooks.AfterSourceLocation(ctx, sl); err != nil {
			return nil, newFailedToRunHookErr(err, "AfterSourceLocation")
		}
	}
	return ctx.Sources, errors.GroupOrNil(errs)
}

func (s sourceLoadingStage) processSourceLocation(ctx PipelineContext, sl string) ([]Source, error) {
	var sources []Source
	for _, l := range s.SourceLoaders {
		if !l.Supports(sl) {
			continue
		}
		loadedSources, err := l.Load(sl)
		if err != nil {
			return nil, err
		}
		ctx.Sources = append(ctx.Sources, loadedSources...)
	}
	return sources, nil
}

type UnitLoadingStageHooksAdapter struct{}

func (_ UnitLoadingStageHooksAdapter) Before(_ PipelineContext) error                 { return nil }
func (_ UnitLoadingStageHooksAdapter) After(_ PipelineContext) error                  { return nil }
func (_ UnitLoadingStageHooksAdapter) BeforeSource(_ PipelineContext, _ Source) error { return nil }
func (_ UnitLoadingStageHooksAdapter) AfterSource(_ PipelineContext, _ Source) error  { return nil }
func (_ UnitLoadingStageHooksAdapter) OnError(_ PipelineContext, err error) error     { return err }

type UnitProcessingStageHooksAdapter struct {
}

func (_ UnitProcessingStageHooksAdapter) Before(_ PipelineContext) error { return nil }
func (_ UnitProcessingStageHooksAdapter) After(_ PipelineContext) error  { return nil }
func (_ UnitProcessingStageHooksAdapter) BeforeProcessor(_ PipelineContext, _ string) error {
	return nil
}
func (_ UnitProcessingStageHooksAdapter) AfterProcessor(_ PipelineContext, _ string) error {
	return nil
}
func (_ UnitProcessingStageHooksAdapter) OnError(_ PipelineContext, err error) error { return err }

type unitLoadingStage struct {
	Loaders []UnitLoader
	Hooks   UnitLoadingStageHooks
}

func (s unitLoadingStage) Run(ctx PipelineContext, sources []Source) ([]Unit, error) {
	if s.Hooks == nil {
		s.Hooks = UnitLoadingStageHooksAdapter{}
	}

	if err := s.Hooks.Before(ctx); err != nil {
		return nil, err
	}

	errs := errors.NewGroup(errors.InternalErrorCode)
	for _, src := range sources {
		if err := ctx.Err(); err != nil {
			return nil, s.handleError(ctx, err)
		}

		if err := s.Hooks.BeforeSource(ctx, src); err != nil {
			return nil, err
		}

		for _, l := range s.Loaders {
			if !l.SupportsSource(src) {
				continue
			}

			loadedUnits, err := l.Load(src)
			if err != nil {
				errs = errs.Append(err)
				continue
			}
			ctx.Units = append(ctx.Units, loadedUnits...)
		}

		if err := s.Hooks.AfterSource(ctx, src); err != nil {
			return nil, err
		}
	}

	if err := s.Hooks.After(ctx); err != nil {
		return nil, err
	}

	return ctx.Units, s.handleError(ctx, errors.GroupOrNil(errs))
}

func (s unitLoadingStage) handleError(ctx PipelineContext, err error) error {
	return s.Hooks.OnError(ctx, err)
}

type unitProcessingStage struct {
	Processors []UnitProcessor
	Hooks      UnitProcessingStageHooks
}

func (s unitProcessingStage) Run(ctx PipelineContext, units []Unit) ([]Artifact, error) {
	if s.Hooks == nil {
		s.Hooks = UnitProcessingStageHooksAdapter{}
	}

	if err := s.Hooks.Before(ctx); err != nil {
		return nil, newFailedToRunHookErr(err, "Before")
	}

	artifacts, err := s.run(ctx, units)
	if err != nil {
		return nil, s.Hooks.OnError(ctx, errors.WrapWithMessage(err, UnitProcessingFailedErrorCode, "failed processing units"))
	}

	if err := s.Hooks.After(ctx); err != nil {
		return nil, newFailedToRunHookErr(err, "After")
	}

	return artifacts, nil
}

func (s unitProcessingStage) run(ctx PipelineContext, units []Unit) ([]Artifact, error) {
	for _, processor := range s.Processors {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if err := s.Hooks.BeforeProcessor(ctx, processor.Name()); err != nil {
			return nil, newFailedToRunHookErr(err, "AfterProcessor")
		}

		artifacts, err := processor.Process(UnitProcessingContext{
			Context:   ctx,
			Units:     units,
			Artifacts: ctx.Artifacts,
		})
		if err != nil {
			return nil, fmt.Errorf("processor %q returned an error :%w", processor.Name(), err)
		}

		ctx.Artifacts = append(ctx.Artifacts, artifacts...)

		if err := s.Hooks.AfterProcessor(ctx, processor.Name()); err != nil {
			return nil, newFailedToRunHookErr(err, "AfterProcessor")
		}
	}

	return ctx.Artifacts, nil
}

type ArtifactProcessingStageHooksAdapter struct {
}

func (_ ArtifactProcessingStageHooksAdapter) Before(PipelineContext) error { return nil }
func (_ ArtifactProcessingStageHooksAdapter) After(PipelineContext) error  { return nil }
func (_ ArtifactProcessingStageHooksAdapter) BeforeProcessor(PipelineContext, string) error {
	return nil
}
func (_ ArtifactProcessingStageHooksAdapter) AfterProcessor(PipelineContext, string) error {
	return nil
}
func (_ ArtifactProcessingStageHooksAdapter) OnError(_ PipelineContext, err error) error { return err }

type artifactProcessingStage struct {
	Registry   ArtifactRegistry
	Processors []ArtifactProcessor
	Hooks      ArtifactProcessingStageHooks
}

func (s artifactProcessingStage) Run(ctx PipelineContext, artifacts []Artifact) error {
	if s.Hooks == nil {
		s.Hooks = ArtifactProcessingStageHooksAdapter{}
	}

	if err := s.Hooks.Before(ctx); err != nil {
		return newFailedToRunHookErr(err, "Before")
	}

	if err := s.run(ctx, artifacts); err != nil {
		err = errors.WrapWithMessage(err, ArtifactProcessingFailedErrorCode, "failed processing artifacts")
		return s.Hooks.OnError(ctx, err)
	}

	if err := s.Hooks.After(ctx); err != nil {
		return newFailedToRunHookErr(err, "After")
	}

	return nil
}

func (s artifactProcessingStage) run(ctx PipelineContext, artifacts []Artifact) (err error) {
	if s.Registry == nil {
		s.Registry = &InMemoryArtifactRegistry{}
	}

	if err := s.Registry.Load(); err != nil {
		return fmt.Errorf("failed loading artifact registry: %w", err)
	}

	defer func() {
		if saveErr := s.Registry.Save(); saveErr != nil {
			saveErr = fmt.Errorf("failed saving artifact registry: %w", err)
			if err != nil {
				err = errors.NewGroup(ArtifactProcessingFailedErrorCode, err, saveErr)
			} else {
				err = saveErr
			}
		}
	}()

	for _, processor := range s.Processors {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.Hooks.BeforeProcessor(ctx, processor.Name()); err != nil {
			return newFailedToRunHookErr(err, "BeforeProcessor")
		}
		if err := s.runProcessor(ctx, processor, artifacts); err != nil {
			return fmt.Errorf("artifact processor %q returned an error: %w", processor.Name(), err)
		}
		if err := s.Hooks.AfterProcessor(ctx, processor.Name()); err != nil {
			return newFailedToRunHookErr(err, "AfterProcessor")
		}
	}

	return nil
}

func (s artifactProcessingStage) runProcessor(ctx PipelineContext, processor ArtifactProcessor, artifacts []Artifact) error {
	processorName := processor.Name()
	apCtx := ArtifactProcessingContext{
		Context:   ctx,
		Units:     ctx.Units,
		Artifacts: artifacts,
		ArtifactRegistry: ProcessorArtifactRegistry{
			processorName: processorName,
			registry:      s.Registry,
		},
		processorName: processorName,
	}

	return processor.Process(apCtx)
}

func newFailedToRunHookErr(err error, hookName string) error {
	return fmt.Errorf("hook %q returned an error: %w", hookName, err)
}
