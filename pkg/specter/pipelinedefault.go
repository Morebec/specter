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
	UnitPreprocessingStage  UnitPreprocessingStage
	UnitProcessingStage     UnitProcessingStage
	ArtifactProcessingStage ArtifactProcessingStage
}

// Run the DefaultPipeline from start to finish.
func (p DefaultPipeline) Run(ctx context.Context, runMode RunMode, sourceLocations []string) (PipelineResult, error) {
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

func (p DefaultPipeline) run(ctx *PipelineContext, sourceLocations []string, runMode RunMode) error {
	if err := p.runSourceLoadingStage(ctx, sourceLocations); err != nil {
		return err
	}
	if runMode == StopAfterSourceLoadingStage {
		return nil
	}

	if err := p.runUnitLoadingStage(ctx); err != nil {
		return err
	}
	if runMode == StopAfterUnitLoadingStage {
		return nil
	}

	if err := p.runUnitPreprocessingStage(ctx); err != nil {
		return err
	}
	if runMode == StopAfterPreprocessingStage {
		return nil
	}

	if err := p.runUnitProcessingStage(ctx); err != nil {
		return err
	}
	if runMode == StopAfterUnitProcessingStage {
		return nil
	}

	if err := p.runArtifactProcessingStage(ctx); err != nil {
		return err
	}
	return nil
}

func (p DefaultPipeline) runSourceLoadingStage(ctx *PipelineContext, sourceLocations []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var err error
	if p.SourceLoadingStage != nil {
		ctx.Sources, err = p.SourceLoadingStage.Run(*ctx, sourceLocations)
		if err != nil {
			return errors.WrapWithMessage(err, SourceLoadingFailedErrorCode, "failed loading sources")
		}
	}

	return nil
}

func (p DefaultPipeline) runUnitLoadingStage(ctx *PipelineContext) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var err error
	if p.UnitLoadingStage != nil {
		ctx.Units, err = p.UnitLoadingStage.Run(*ctx, ctx.Sources)
		if err != nil {
			return errors.WrapWithMessage(err, UnitLoadingFailedErrorCode, "failed loading units")
		}
	}

	return nil
}

func (p DefaultPipeline) runUnitProcessingStage(ctx *PipelineContext) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if p.UnitProcessingStage != nil {
		var err error
		ctx.Artifacts, err = p.UnitProcessingStage.Run(*ctx, ctx.Units)
		if err != nil {
			return errors.WrapWithMessage(err, UnitProcessingFailedErrorCode, "failed processing units")
		}
	}
	return nil
}

func (p DefaultPipeline) runArtifactProcessingStage(ctx *PipelineContext) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if p.ArtifactProcessingStage != nil {
		if err := p.ArtifactProcessingStage.Run(*ctx, ctx.Artifacts); err != nil {
			return errors.WrapWithMessage(err, ArtifactProcessingFailedErrorCode, "failed processing artifacts")
		}
	}
	return nil
}

func (p DefaultPipeline) runUnitPreprocessingStage(ctx *PipelineContext) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if p.UnitPreprocessingStage != nil {
		var err error
		ctx.Units, err = p.UnitPreprocessingStage.Run(*ctx, ctx.Units)
		if err != nil {
			return errors.WrapWithMessage(err, UnitPreprocessingFailedErrorCode, "failed preprocessing units")
		}
	}
	return nil
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

func (s sourceLoadingStage) processSourceLocation(_ PipelineContext, sl string) ([]Source, error) {
	var sources []Source
	for _, l := range s.SourceLoaders {
		if !l.Supports(sl) {
			continue
		}
		loadedSources, err := l.Load(sl)
		if err != nil {
			return nil, err
		}
		sources = append(sources, loadedSources...)
	}
	return sources, nil
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

	units, err := s.run(ctx, sources)
	if err != nil {
		return units, s.Hooks.OnError(ctx, errors.WrapWithMessage(err, UnitLoadingFailedErrorCode, "failed to load units"))
	}

	if err := s.Hooks.After(ctx); err != nil {
		return nil, err
	}

	return units, nil
}

func (s unitLoadingStage) run(ctx PipelineContext, sources []Source) ([]Unit, error) {
	var units []Unit
	for _, src := range sources {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if err := s.Hooks.BeforeSource(ctx, src); err != nil {
			return nil, err
		}

		uns, err := s.runLoader(src)
		if err != nil {
			return nil, err
		}
		units = append(units, uns...)

		if err := s.Hooks.AfterSource(ctx, src); err != nil {
			return nil, err
		}
	}
	return units, nil
}

func (s unitLoadingStage) runLoader(src Source) ([]Unit, error) {
	var units []Unit
	for _, l := range s.Loaders {
		if !l.SupportsSource(src) {
			continue
		}

		loadedUnits, err := l.Load(src)
		if err != nil {
			return nil, err
		}
		units = append(units, loadedUnits...)
	}
	return units, nil
}

type UnitLoadingStageHooksAdapter struct{}

func (_ UnitLoadingStageHooksAdapter) Before(_ PipelineContext) error                 { return nil }
func (_ UnitLoadingStageHooksAdapter) After(_ PipelineContext) error                  { return nil }
func (_ UnitLoadingStageHooksAdapter) BeforeSource(_ PipelineContext, _ Source) error { return nil }
func (_ UnitLoadingStageHooksAdapter) AfterSource(_ PipelineContext, _ Source) error  { return nil }
func (_ UnitLoadingStageHooksAdapter) OnError(_ PipelineContext, err error) error     { return err }

type unitPreprocessingStage struct {
	Preprocessors []UnitPreprocessor
	Hooks         UnitPreprocessingStageHooks
}

func (s unitPreprocessingStage) Run(ctx PipelineContext, units []Unit) ([]Unit, error) {
	if s.Hooks == nil {
		s.Hooks = UnitPreprocessingStageHooksAdapter{}
	}

	if err := s.Hooks.Before(ctx); err != nil {
		return nil, newFailedToRunHookErr(err, "Before")
	}

	units, err := s.run(ctx, units)
	if err != nil {
		return nil, s.Hooks.OnError(ctx, errors.WrapWithMessage(err, UnitProcessingFailedErrorCode, "failed preprocessing units"))
	}

	if err := s.Hooks.After(ctx); err != nil {
		return nil, newFailedToRunHookErr(err, "After")
	}

	return units, nil
}

func (s unitPreprocessingStage) run(ctx PipelineContext, units []Unit) ([]Unit, error) {
	var err error
	for _, p := range s.Preprocessors {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if err := s.Hooks.BeforePreprocessor(ctx, p.Name()); err != nil {
			return nil, newFailedToRunHookErr(err, "BeforePreprocessor")
		}

		units, err = p.Preprocess(ctx, units)
		if err != nil {
			return nil, fmt.Errorf("preprocessor %q returned an error: %w", p.Name(), err)
		}
		ctx.Units = units

		if err := s.Hooks.AfterPreprocessor(ctx, p.Name()); err != nil {
			return nil, newFailedToRunHookErr(err, "AfterPreprocessor")
		}
	}

	return units, nil
}

type UnitPreprocessingStageHooksAdapter struct {
}

func (u UnitPreprocessingStageHooksAdapter) Before(PipelineContext) error { return nil }

func (u UnitPreprocessingStageHooksAdapter) After(PipelineContext) error { return nil }

func (u UnitPreprocessingStageHooksAdapter) BeforePreprocessor(PipelineContext, string) error {
	return nil
}

func (u UnitPreprocessingStageHooksAdapter) AfterPreprocessor(PipelineContext, string) error {
	return nil
}

func (u UnitPreprocessingStageHooksAdapter) OnError(_ PipelineContext, err error) error { return err }

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
			return nil, fmt.Errorf("processor %q returned an error: %w", processor.Name(), err)
		}

		ctx.Artifacts = append(ctx.Artifacts, artifacts...)

		if err := s.Hooks.AfterProcessor(ctx, processor.Name()); err != nil {
			return nil, newFailedToRunHookErr(err, "AfterProcessor")
		}
	}

	return ctx.Artifacts, nil
}

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
