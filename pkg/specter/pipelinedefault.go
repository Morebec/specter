package specter

import (
	"context"
	"fmt"
	"github.com/morebec/go-errors/errors"
)

// DefaultPipeline is the service responsible to run a specter DefaultPipeline.
type DefaultPipeline struct {
	TimeProvider TimeProvider

	sourceLoadingStage      sourceLoadingStage
	unitLoadingStage        unitLoadingStage
	unitProcessingStage     unitProcessingStage
	artifactProcessingStage artifactProcessingStage
}

// Run the DefaultPipeline from start to finish.
func (p DefaultPipeline) Run(ctx context.Context, sourceLocations []string, runMode RunMode) (PipelineResult, error) {
	pctx := &PipelineContext{
		Context: ctx,
		pipelineContext: pipelineContext{
			StartedAt:       p.TimeProvider(),
			SourceLocations: sourceLocations,
			RunMode:         runMode,
		},
	}

	err := p.run(pctx, sourceLocations)

	result := PipelineResult{
		pipelineContext: pctx.pipelineContext,
		EndedAt:         p.TimeProvider(),
	}

	return result, err
}

func (p DefaultPipeline) run(pctx *PipelineContext, sourceLocations []string) error {
	var err error

	pctx.Sources, err = p.sourceLoadingStage.Run(*pctx, sourceLocations)
	if err != nil {
		return errors.WrapWithMessage(err, SourceLoadingFailedErrorCode, "failed loading sources")
	}

	pctx.Units, err = p.unitLoadingStage.Run(*pctx, pctx.Sources)
	if err != nil {
		return errors.WrapWithMessage(err, UnitLoadingFailedErrorCode, "failed loading units")
	}

	pctx.Artifacts, err = p.unitProcessingStage.Run(*pctx, pctx.Units)
	if err != nil {
		return errors.WrapWithMessage(err, UnitProcessingFailedErrorCode, "failed processing units")
	}

	if err := p.artifactProcessingStage.Run(*pctx, pctx.Artifacts); err != nil {
		return errors.WrapWithMessage(err, UnitProcessingFailedErrorCode, "failed processing artifacts")
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

	ctx.SourceLocations = sourceLocations
	if err := s.Hooks.Before(ctx); err != nil {
		return nil, err
	}

	errs := errors.NewGroup(SourceLoadingFailedErrorCode)

	for _, sl := range sourceLocations {
		if err := ctx.Err(); err != nil {
			return nil, s.HandleError(ctx, err)
		}

		if err := s.Hooks.BeforeSourceLocation(ctx, sl); err != nil {
			return nil, err
		}

		for _, l := range s.SourceLoaders {
			if !l.Supports(sl) {
				continue
			}

			loadedSources, err := l.Load(sl)
			if err != nil {
				errs = errs.Append(err)
				continue
			}
			ctx.Sources = append(ctx.Sources, loadedSources...)
		}

		if err := s.Hooks.AfterSourceLocation(ctx, sl); err != nil {
			return nil, err
		}
	}

	if err := s.Hooks.After(ctx); err != nil {
		return nil, err
	}

	return ctx.Sources, s.HandleError(ctx, errors.GroupOrNil(errs))
}

func (s sourceLoadingStage) HandleError(ctx PipelineContext, err error) error {
	return s.Hooks.OnError(ctx, err)
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
		return nil, err
	}

	for _, processor := range s.Processors {
		if err := ctx.Err(); err != nil {
			return nil, s.handleError(ctx, err)
		}

		if err := s.Hooks.BeforeProcessor(ctx, processor.Name()); err != nil {
			return nil, err
		}

		artifacts, err := processor.Process(ProcessingContext{
			Context:   ctx,
			Units:     units,
			Artifacts: nil,
		})
		if err != nil {
			return nil, s.handleError(ctx, fmt.Errorf("failed to run processor %q: %w", processor.Name(), err))
		}

		ctx.Artifacts = append(ctx.Artifacts, artifacts...)
	}

	if err := s.Hooks.After(ctx); err != nil {
		return nil, err
	}

	return ctx.Artifacts, nil
}

func (s unitProcessingStage) handleError(ctx PipelineContext, err error) error {
	return s.Hooks.OnError(ctx, err)
}

type ArtifactProcessingStageHooksAdapter struct {
}

func (_ ArtifactProcessingStageHooksAdapter) Before(_ PipelineContext) error { return nil }
func (_ ArtifactProcessingStageHooksAdapter) After(_ PipelineContext) error  { return nil }
func (_ ArtifactProcessingStageHooksAdapter) BeforeProcessor(_ PipelineContext, _ string) error {
	return nil
}
func (_ ArtifactProcessingStageHooksAdapter) AfterProcessor(_ PipelineContext, _ string) error {
	return nil
}
func (_ ArtifactProcessingStageHooksAdapter) OnError(_ PipelineContext, err error) error { return err }

type artifactProcessingStage struct {
	ArtifactRegistry   ArtifactRegistry
	ArtifactProcessors []ArtifactProcessor
	Hooks              ArtifactProcessingStageHooks
}

func (s artifactProcessingStage) Run(ctx PipelineContext, artifacts []Artifact) (err error) {
	if s.ArtifactRegistry == nil {
		s.ArtifactRegistry = &InMemoryArtifactRegistry{}
	}

	if s.Hooks == nil {
		s.Hooks = ArtifactProcessingStageHooksAdapter{}
	}

	if err := s.ArtifactRegistry.Load(); err != nil {
		return errors.WrapWithMessage(
			fmt.Errorf("failed loading artifact registry: %w", err),
			ArtifactProcessingFailedErrorCode,
			"failed processing artifacts",
		)
	}

	defer func() {
		if saveErr := s.ArtifactRegistry.Save(); saveErr != nil {
			saveErr = errors.WrapWithMessage(
				fmt.Errorf("failed saving artifact registry: %w", err),
				ArtifactProcessingFailedErrorCode,
				"failed processing artifacts",
			)
			if err != nil {
				err = errors.NewGroup(ArtifactProcessingFailedErrorCode, err, saveErr)
			} else {
				err = saveErr
			}
		}
	}()

	for _, processor := range s.ArtifactProcessors {
		if err := ctx.Err(); err != nil {
			return errors.WrapWithMessage(err, ArtifactProcessingFailedErrorCode, "failed processing artifacts")
		}

		processorName := processor.Name()
		apCtx := ArtifactProcessingContext{
			Context:   ctx,
			Units:     ctx.Units,
			Artifacts: artifacts,
			ArtifactRegistry: ProcessorArtifactRegistry{
				processorName: processorName,
				registry:      s.ArtifactRegistry,
			},
			processorName: processorName,
		}

		if err := processor.Process(apCtx); err != nil {
			return errors.WrapWithMessage(
				err,
				ArtifactProcessingFailedErrorCode,
				fmt.Sprintf("failed processing artifacts with processor %q", processorName),
			)
		}
	}

	return nil
}
