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

package specter_test

import (
	"context"
	"github.com/morebec/specter/pkg/specter"
	"github.com/morebec/specter/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSourceLoadingStageHooksAdapter(t *testing.T) {
	t.Run("Before should not return error", func(t *testing.T) {
		a := specter.SourceLoadingStageHooksAdapter{}
		err := a.Before(specter.PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("After should not return error", func(t *testing.T) {
		a := specter.SourceLoadingStageHooksAdapter{}
		err := a.After(specter.PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("BeforeSourceLocation should not return error", func(t *testing.T) {
		a := specter.SourceLoadingStageHooksAdapter{}
		err := a.BeforeSourceLocation(specter.PipelineContext{}, "")
		require.NoError(t, err)
	})

	t.Run("AfterSourceLocation should not return error", func(t *testing.T) {
		a := specter.SourceLoadingStageHooksAdapter{}
		err := a.AfterSourceLocation(specter.PipelineContext{}, "")
		require.NoError(t, err)
	})

	t.Run("OnError should return error", func(t *testing.T) {
		a := specter.SourceLoadingStageHooksAdapter{}
		err := a.OnError(specter.PipelineContext{}, assert.AnError)
		require.Equal(t, assert.AnError, err)
	})
}

func TestUnitLoadingStageHooksAdapter(t *testing.T) {
	t.Run("Before should not return error", func(t *testing.T) {

		a := specter.UnitLoadingStageHooksAdapter{}
		err := a.Before(specter.PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("After should not return error", func(t *testing.T) {

		a := specter.UnitLoadingStageHooksAdapter{}
		err := a.After(specter.PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("BeforeSource should not return error", func(t *testing.T) {

		a := specter.UnitLoadingStageHooksAdapter{}
		err := a.BeforeSource(specter.PipelineContext{}, specter.Source{})
		require.NoError(t, err)
	})

	t.Run("AfterSource should not return error", func(t *testing.T) {

		a := specter.UnitLoadingStageHooksAdapter{}
		err := a.AfterSource(specter.PipelineContext{}, specter.Source{})
		require.NoError(t, err)
	})

	t.Run("OnError should return error", func(t *testing.T) {

		a := specter.UnitLoadingStageHooksAdapter{}
		err := a.OnError(specter.PipelineContext{}, assert.AnError)
		require.Equal(t, assert.AnError, err)
	})
}

func TestUnitProcessingStageHooksAdapter(t *testing.T) {
	t.Run("Before should not return error", func(t *testing.T) {
		a := specter.UnitProcessingStageHooksAdapter{}
		err := a.Before(specter.PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("After should not return error", func(t *testing.T) {
		a := specter.UnitProcessingStageHooksAdapter{}
		err := a.After(specter.PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("BeforeProcessor should not return error", func(t *testing.T) {
		a := specter.UnitProcessingStageHooksAdapter{}
		err := a.BeforeProcessor(specter.PipelineContext{}, "processor")
		require.NoError(t, err)
	})

	t.Run("AfterProcessor should not return error", func(t *testing.T) {
		a := specter.UnitProcessingStageHooksAdapter{}
		err := a.AfterProcessor(specter.PipelineContext{}, "processor")
		require.NoError(t, err)
	})

	t.Run("OnError should return error", func(t *testing.T) {
		a := specter.UnitProcessingStageHooksAdapter{}
		err := a.OnError(specter.PipelineContext{}, assert.AnError)
		require.Equal(t, assert.AnError, err)
	})
}

func TestArtifactProcessingStageHooksAdapter(t *testing.T) {
	t.Run("Before should not return error", func(t *testing.T) {
		a := specter.ArtifactProcessingStageHooksAdapter{}
		err := a.Before(specter.PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("After should not return error", func(t *testing.T) {
		a := specter.ArtifactProcessingStageHooksAdapter{}
		err := a.After(specter.PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("BeforeProcessor should not return error", func(t *testing.T) {
		a := specter.ArtifactProcessingStageHooksAdapter{}
		err := a.BeforeProcessor(specter.PipelineContext{}, "processor")
		require.NoError(t, err)
	})

	t.Run("AfterProcessor should not return error", func(t *testing.T) {
		a := specter.ArtifactProcessingStageHooksAdapter{}
		err := a.AfterProcessor(specter.PipelineContext{}, "processor")
		require.NoError(t, err)
	})

	t.Run("OnError should return error", func(t *testing.T) {
		a := specter.ArtifactProcessingStageHooksAdapter{}
		err := a.OnError(specter.PipelineContext{}, assert.AnError)
		require.Equal(t, assert.AnError, err)
	})
}

func TestDefaultPipeline_Run(t *testing.T) {
	currentTime := time.Date(2024, 01, 01, 0, 0, 0, 0, time.UTC)

	type given struct {
		SourceLoadingStage      specter.SourceLoadingStage
		UnitLoadingStage        specter.UnitLoadingStage
		UnitProcessingStage     specter.UnitProcessingStage
		ArtifactProcessingStage specter.ArtifactProcessingStage
	}
	type args struct {
		ctx             context.Context
		sourceLocations []string
		runMode         specter.RunMode
	}
	type then struct {
		expectedResult specter.PipelineResult
		expectedError  require.ErrorAssertionFunc
	}
	tests := []struct {
		name  string
		given given
		when  args
		then  then
	}{
		{
			name: "WHEN an empty RunMode is provided THEN should default to RunThrough",
			when: args{
				ctx:             context.Background(),
				sourceLocations: nil,
				runMode:         "",
			},
			then: then{
				expectedResult: specter.PipelineResult{
					PipelineContextData: specter.PipelineContextData{
						StartedAt: currentTime,
						RunMode:   specter.RunThrough,
					},
					EndedAt: currentTime,
				},
				expectedError: require.NoError,
			},
		},
		{
			// Empty pipeline is ok.
			name: "GIVEN nil stages provided THEN no error should be returned",
			given: given{
				SourceLoadingStage:      nil,
				UnitLoadingStage:        nil,
				UnitProcessingStage:     nil,
				ArtifactProcessingStage: nil,
			},
			when: args{
				ctx:             context.Background(),
				sourceLocations: nil,
				runMode:         "some-mode",
			},
			then: then{
				expectedResult: specter.PipelineResult{
					PipelineContextData: specter.PipelineContextData{
						StartedAt: currentTime,
						RunMode:   "some-mode",
					},
					EndedAt: currentTime,
				},
				expectedError: require.NoError,
			},
		},
		// Stage Errors
		{
			name: "GIVEN source loading stage fails THEN an error should be returned",
			given: given{
				SourceLoadingStage: FailingSourceLoadingStage{},
			},
			when: args{
				ctx:     context.Background(),
				runMode: specter.RunThrough,
			},
			then: then{
				expectedResult: specter.PipelineResult{
					PipelineContextData: specter.PipelineContextData{
						StartedAt: currentTime,
						RunMode:   specter.RunThrough,
					},
					EndedAt: currentTime,
				},
				expectedError: testutils.RequireErrorWithCode(specter.SourceLoadingFailedErrorCode),
			},
		},
		{
			name: "GIVEN unit loading stage fails THEN an error should be returned",
			given: given{
				UnitLoadingStage: FailingUnitLoadingStage{},
			},
			when: args{
				ctx:     context.Background(),
				runMode: specter.RunThrough,
			},
			then: then{
				expectedResult: specter.PipelineResult{
					PipelineContextData: specter.PipelineContextData{
						StartedAt: currentTime,
						RunMode:   specter.RunThrough,
					},
					EndedAt: currentTime,
				},
				expectedError: testutils.RequireErrorWithCode(specter.UnitLoadingFailedErrorCode),
			},
		},
		{
			name: "GIVEN unit processing stage fails THEN an error should be returned",
			given: given{
				UnitProcessingStage: FailingUnitProcessingStage{},
			},
			when: args{
				ctx:     context.Background(),
				runMode: specter.RunThrough,
			},
			then: then{
				expectedResult: specter.PipelineResult{
					PipelineContextData: specter.PipelineContextData{
						StartedAt: currentTime,
						RunMode:   specter.RunThrough,
					},
					EndedAt: currentTime,
				},
				expectedError: testutils.RequireErrorWithCode(specter.UnitProcessingFailedErrorCode),
			},
		},
		{
			name: "GIVEN artifact processing stage fails THEN an error should be returned",
			given: given{
				ArtifactProcessingStage: FailingArtifactProcessingStage{},
			},
			when: args{
				ctx:     context.Background(),
				runMode: specter.RunThrough,
			},
			then: then{
				expectedResult: specter.PipelineResult{
					PipelineContextData: specter.PipelineContextData{
						StartedAt: currentTime,
						RunMode:   specter.RunThrough,
					},
					EndedAt: currentTime,
				},
				expectedError: testutils.RequireErrorWithCode(specter.ArtifactProcessingFailedErrorCode),
			},
		},

		// Run Modes
		{
			name: "WHEN stop after source loading THEN it should stop and no error should be returned",
			given: given{
				UnitLoadingStage: FailingUnitLoadingStage{}, // Should not fail
			},
			when: args{
				ctx:     context.Background(),
				runMode: specter.StopAfterSourceLoadingStage,
			},
			then: then{
				expectedResult: specter.PipelineResult{
					PipelineContextData: specter.PipelineContextData{
						StartedAt: currentTime,
						RunMode:   specter.StopAfterSourceLoadingStage,
					},
					EndedAt: currentTime,
				},
				expectedError: require.NoError,
			},
		},
		{
			name: "WHEN stop after unit loading THEN it should stop and no error should be returned",
			given: given{
				UnitProcessingStage: FailingUnitProcessingStage{},
			},
			when: args{
				ctx:     context.Background(),
				runMode: specter.StopAfterUnitLoadingStage,
			},
			then: then{
				expectedResult: specter.PipelineResult{
					PipelineContextData: specter.PipelineContextData{
						StartedAt: currentTime,
						RunMode:   specter.StopAfterUnitLoadingStage,
					},
					EndedAt: currentTime,
				},
				expectedError: require.NoError,
			},
		},
		{
			name: "WHEN stop after unit processing THEN it should stop and no error should be returned",
			given: given{
				ArtifactProcessingStage: FailingArtifactProcessingStage{},
			},
			when: args{
				ctx:     context.Background(),
				runMode: specter.StopAfterUnitProcessingStage,
			},
			then: then{
				expectedResult: specter.PipelineResult{
					PipelineContextData: specter.PipelineContextData{
						StartedAt: currentTime,
						RunMode:   specter.StopAfterUnitProcessingStage,
					},
					EndedAt: currentTime,
				},
				expectedError: require.NoError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := specter.DefaultPipeline{
				TimeProvider:            staticTimeProvider(currentTime),
				SourceLoadingStage:      tt.given.SourceLoadingStage,
				UnitLoadingStage:        tt.given.UnitLoadingStage,
				UnitProcessingStage:     tt.given.UnitProcessingStage,
				ArtifactProcessingStage: tt.given.ArtifactProcessingStage,
			}
			got, err := p.Run(tt.when.ctx, tt.when.sourceLocations, tt.when.runMode)
			tt.then.expectedError(t, err)
			assert.Equal(t, tt.then.expectedResult, got)
		})
	}
}

func Test_sourceLoadingStage_Run(t *testing.T) {
	t.Run("should call all hooks under normal processing", func(t *testing.T) {
		recorder := sourceLoadingStageHooksCallRecorder{}

		stage := specter.DefaultSourceLoadingStage{
			SourceLoaders: []specter.SourceLoader{
				specter.FunctionalSourceLoader{
					SupportsFunc: func(string) bool {
						return true
					},
					LoadFunc: func(location string) ([]specter.Source, error) {
						return nil, nil
					},
				},
			},
			Hooks: &recorder,
		}

		units, err := stage.Run(specter.PipelineContext{Context: context.Background()}, []string{
			"/path/to/file",
		})
		require.NoError(t, err)
		require.Nil(t, units)

		assert.True(t, recorder.beforeCalled)
		assert.True(t, recorder.beforeSourceLocationCalled)
		assert.True(t, recorder.afterSourceLocationCalled)
		assert.True(t, recorder.afterCalled)
	})

	t.Run("should call hooks until error", func(t *testing.T) {
		recorder := sourceLoadingStageHooksCallRecorder{}

		stage := specter.DefaultSourceLoadingStage{
			SourceLoaders: []specter.SourceLoader{
				specter.FunctionalSourceLoader{
					SupportsFunc: func(string) bool {
						return true
					},
					LoadFunc: func(location string) ([]specter.Source, error) {
						return nil, assert.AnError
					},
				},
			},
			Hooks: &recorder,
		}

		units, err := stage.Run(specter.PipelineContext{Context: context.Background()}, []string{
			"/path/to/file",
		})
		require.Error(t, err)
		require.Nil(t, units)

		assert.True(t, recorder.beforeCalled)
		assert.True(t, recorder.beforeSourceLocationCalled)
		assert.True(t, recorder.onErrorCalled)
		assert.False(t, recorder.afterSourceLocationCalled)
		assert.False(t, recorder.afterCalled)
	})
}

func Test_unitProcessingStage_Run(t *testing.T) {
	t.Run("should call all hooks under normal processing", func(t *testing.T) {
		recorder := unitProcessingStageHooksCallRecorder{}

		stage := specter.DefaultUnitProcessingStage{
			Processors: []specter.UnitProcessor{
				specter.NewUnitProcessorFunc("", func(specter.UnitProcessingContext) ([]specter.Artifact, error) {
					return nil, nil
				}),
			},
			Hooks: &recorder,
		}

		units, err := stage.Run(specter.PipelineContext{Context: context.Background()}, nil)
		require.NoError(t, err)
		require.Nil(t, units)

		assert.True(t, recorder.beforeCalled)
		assert.True(t, recorder.beforeProcessorCalled)
		assert.True(t, recorder.afterProcessorCalled)
		assert.True(t, recorder.afterCalled)
	})

	t.Run("should call hooks until error", func(t *testing.T) {
		recorder := unitProcessingStageHooksCallRecorder{}

		stage := specter.DefaultUnitProcessingStage{
			Processors: []specter.UnitProcessor{
				specter.NewUnitProcessorFunc("", func(specter.UnitProcessingContext) ([]specter.Artifact, error) {
					return nil, assert.AnError
				}),
			},
			Hooks: &recorder,
		}

		units, err := stage.Run(specter.PipelineContext{Context: context.Background()}, nil)
		require.Error(t, err)
		require.Nil(t, units)

		assert.True(t, recorder.beforeCalled)
		assert.True(t, recorder.beforeProcessorCalled)
		assert.True(t, recorder.onErrorCalled)
		assert.False(t, recorder.afterProcessorCalled)
		assert.False(t, recorder.afterCalled)
	})
}

func Test_artifactProcessingStage_Run(t *testing.T) {
	t.Run("should call all hooks under normal processing", func(t *testing.T) {
		recorder := artifactProcessingStageHooksCallRecorder{}

		stage := specter.DefaultArtifactProcessingStage{
			Processors: []specter.ArtifactProcessor{
				specter.NewArtifactProcessorFunc("", func(ctx specter.ArtifactProcessingContext) error { return nil }),
			},
			Hooks: &recorder,
		}

		err := stage.Run(specter.PipelineContext{Context: context.Background()}, []specter.Artifact{
			&specter.FileArtifact{},
		})
		require.NoError(t, err)

		assert.True(t, recorder.beforeCalled)
		assert.True(t, recorder.beforeProcessorCalled)
		assert.True(t, recorder.afterProcessorCalled)
		assert.True(t, recorder.afterCalled)
	})

	t.Run("should call hooks until error", func(t *testing.T) {
		recorder := artifactProcessingStageHooksCallRecorder{}

		stage := specter.DefaultArtifactProcessingStage{
			Processors: []specter.ArtifactProcessor{
				specter.NewArtifactProcessorFunc("", func(ctx specter.ArtifactProcessingContext) error { return assert.AnError }),
			},
			Hooks: &recorder,
		}

		err := stage.Run(specter.PipelineContext{Context: context.Background()}, []specter.Artifact{
			&specter.FileArtifact{},
		})
		require.Error(t, err)

		assert.True(t, recorder.beforeCalled)
		assert.True(t, recorder.beforeProcessorCalled)
		assert.True(t, recorder.onErrorCalled)
		assert.False(t, recorder.afterProcessorCalled)
		assert.False(t, recorder.afterCalled)
	})
}

type FailingSourceLoadingStage struct{}

func (f FailingSourceLoadingStage) Run(specter.PipelineContext, []string) ([]specter.Source, error) {
	return nil, assert.AnError
}

type FailingUnitLoadingStage struct{}

func (f FailingUnitLoadingStage) Run(specter.PipelineContext, []specter.Source) ([]specter.Unit, error) {
	return nil, assert.AnError
}

type FailingUnitProcessingStage struct{}

func (f FailingUnitProcessingStage) Run(specter.PipelineContext, []specter.Unit) ([]specter.Artifact, error) {
	return nil, assert.AnError
}

type FailingArtifactProcessingStage struct{}

func (f FailingArtifactProcessingStage) Run(specter.PipelineContext, []specter.Artifact) error {
	return assert.AnError
}

type sourceLoadingStageHooksCallRecorder struct {
	beforeCalled               bool
	afterCalled                bool
	beforeSourceLocationCalled bool
	afterSourceLocationCalled  bool
	onErrorCalled              bool
}

func (s *sourceLoadingStageHooksCallRecorder) Before(specter.PipelineContext) error {
	s.beforeCalled = true
	return nil
}

func (s *sourceLoadingStageHooksCallRecorder) After(specter.PipelineContext) error {
	s.afterCalled = true
	return nil
}

func (s *sourceLoadingStageHooksCallRecorder) BeforeSourceLocation(specter.PipelineContext, string) error {
	s.beforeSourceLocationCalled = true
	return nil
}

func (s *sourceLoadingStageHooksCallRecorder) AfterSourceLocation(specter.PipelineContext, string) error {
	s.afterSourceLocationCalled = true
	return nil
}

func (s *sourceLoadingStageHooksCallRecorder) OnError(_ specter.PipelineContext, err error) error {
	s.onErrorCalled = true
	return err
}

type unitProcessingStageHooksCallRecorder struct {
	beforeCalled          bool
	afterCalled           bool
	beforeProcessorCalled bool
	afterProcessorCalled  bool
	onErrorCalled         bool
}

func (a *unitProcessingStageHooksCallRecorder) Before(specter.PipelineContext) error {
	a.beforeCalled = true
	return nil
}

func (a *unitProcessingStageHooksCallRecorder) After(specter.PipelineContext) error {
	a.afterCalled = true
	return nil
}

func (a *unitProcessingStageHooksCallRecorder) BeforeProcessor(specter.PipelineContext, string) error {
	a.beforeProcessorCalled = true
	return nil
}

func (a *unitProcessingStageHooksCallRecorder) AfterProcessor(specter.PipelineContext, string) error {
	a.afterProcessorCalled = true
	return nil
}

func (a *unitProcessingStageHooksCallRecorder) OnError(_ specter.PipelineContext, err error) error {
	a.onErrorCalled = true
	return err
}

type artifactProcessingStageHooksCallRecorder struct {
	beforeCalled          bool
	afterCalled           bool
	beforeProcessorCalled bool
	afterProcessorCalled  bool
	onErrorCalled         bool
}

func (a *artifactProcessingStageHooksCallRecorder) Before(specter.PipelineContext) error {
	a.beforeCalled = true
	return nil
}

func (a *artifactProcessingStageHooksCallRecorder) After(specter.PipelineContext) error {
	a.afterCalled = true
	return nil
}

func (a *artifactProcessingStageHooksCallRecorder) BeforeProcessor(specter.PipelineContext, string) error {
	a.beforeProcessorCalled = true
	return nil
}

func (a *artifactProcessingStageHooksCallRecorder) AfterProcessor(specter.PipelineContext, string) error {
	a.afterProcessorCalled = true
	return nil
}

func (a *artifactProcessingStageHooksCallRecorder) OnError(_ specter.PipelineContext, err error) error {
	a.onErrorCalled = true
	return err
}
