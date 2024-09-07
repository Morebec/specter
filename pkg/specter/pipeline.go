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
	"time"
)

type RunMode string

const (
	RunThrough                   RunMode = "run-through"
	StopAfterSourceLoadingStage  RunMode = "stop-after-source-loading-stage"
	StopAfterUnitLoadingStage    RunMode = "stop-after-unit-loading-stage"
	StopAfterUnitProcessingStage RunMode = "stop-after-unit-processing-stage"
)

const SourceLoadingFailedErrorCode = "specter.source_loading_failed"
const UnitLoadingFailedErrorCode = "specter.unit_loading_failed"
const UnitProcessingFailedErrorCode = "specter.unit_processing_failed"
const ArtifactProcessingFailedErrorCode = "specter.artifact_processing_failed"

type PipelineResult struct {
	PipelineContextData
	EndedAt time.Time
}

func (r PipelineResult) ExecutionTime() time.Duration {
	return r.EndedAt.Sub(r.StartedAt)
}

type PipelineContext struct {
	context.Context
	PipelineContextData
}

type PipelineContextData struct {
	StartedAt       time.Time
	SourceLocations []string
	Sources         []Source
	Units           []Unit
	Artifacts       []Artifact
	RunMode         RunMode
}

type Pipeline interface {
	Run(ctx context.Context, sourceLocations []string, runMode RunMode) (result PipelineResult, err error)
}

type SourceLoadingStage interface {
	Run(ctx PipelineContext, sourceLocations []string) ([]Source, error)
}

type UnitLoadingStage interface {
	Run(ctx PipelineContext, sources []Source) ([]Unit, error)
}

type UnitProcessingStage interface {
	Run(ctx PipelineContext, units []Unit) ([]Artifact, error)
}

type ArtifactProcessingStage interface {
	Run(ctx PipelineContext, artifacts []Artifact) error
}

type SourceLoadingStageHooks interface {
	Before(ctx PipelineContext) error
	After(ctx PipelineContext) error
	BeforeSourceLocation(ctx PipelineContext, sourceLocation string) error
	AfterSourceLocation(ctx PipelineContext, sourceLocation string) error
	OnError(ctx PipelineContext, err error) error
}

type UnitLoadingStageHooks interface {
	Before(ctx PipelineContext) error
	After(ctx PipelineContext) error
	BeforeSource(ctx PipelineContext, source Source) error
	AfterSource(ctx PipelineContext, source Source) error
	OnError(ctx PipelineContext, err error) error
}

type UnitProcessingStageHooks interface {
	Before(ctx PipelineContext) error
	After(ctx PipelineContext) error
	BeforeProcessor(ctx PipelineContext, processorName string) error
	AfterProcessor(ctx PipelineContext, processorName string) error
	OnError(ctx PipelineContext, err error) error
}

type ArtifactProcessingStageHooks interface {
	Before(ctx PipelineContext) error
	After(ctx PipelineContext) error
	BeforeProcessor(ctx PipelineContext, processorName string) error
	AfterProcessor(ctx PipelineContext, processorName string) error
	OnError(ctx PipelineContext, err error) error
}
