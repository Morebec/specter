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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRunResult_ExecutionTime(t *testing.T) {
	r := RunResult{}
	r.StartedAt = time.Date(2024, 01, 01, 0, 0, 0, 0, time.UTC)
	r.EndedAt = time.Date(2024, 01, 01, 1, 0, 0, 0, time.UTC)

	require.Equal(t, r.ExecutionTime(), time.Hour*1)
}

func TestSpecter_Run(t *testing.T) {
	testDay := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	type given struct {
		specter func() *Specter
	}

	type when struct {
		context         context.Context
		sourceLocations []string
		executionMode   RunMode
	}

	type then struct {
		expectedRunResult RunResult
		expectedError     assert.ErrorAssertionFunc
	}

	tests := []struct {
		name  string
		given given
		when  when
		then  then
	}{
		{
			name: "WHEN no source locations provided THEN return with no error",
			given: given{
				specter: func() *Specter {
					return New(
						WithTimeProvider(staticTimeProvider(testDay)),
					)
				},
			},
			when: when{
				context:         context.Background(),
				sourceLocations: nil,
				executionMode:   PreviewMode,
			},
			then: then{
				expectedRunResult: RunResult{
					RunMode:        PreviewMode,
					Sources:        nil,
					Specifications: nil,
					Artifacts:      nil,
					StartedAt:      testDay,
					EndedAt:        testDay,
				},
				expectedError: assert.NoError,
			},
		},
		{
			name: "WHEN no execution mode provided THEN assume Preview mode",
			given: given{
				specter: func() *Specter {
					return New(
						WithTimeProvider(staticTimeProvider(testDay)),
					)
				},
			},
			when: when{
				context:         context.Background(),
				sourceLocations: nil,
				executionMode:   "", // No execution mode should default to preview
			},
			then: then{
				expectedRunResult: RunResult{
					RunMode:        PreviewMode,
					Sources:        nil,
					Specifications: nil,
					Artifacts:      nil,
					StartedAt:      testDay,
					EndedAt:        testDay,
				},
				expectedError: assert.NoError,
			},
		},
		{
			name: "WHEN no context is provided THEN assume a context.Background and do not fail",
			given: given{
				specter: func() *Specter {
					return New(
						WithTimeProvider(staticTimeProvider(testDay)),
					)
				},
			},
			when: when{
				context:         nil,
				sourceLocations: nil,
				executionMode:   "", // No execution mode should default to preview
			},
			then: then{
				expectedRunResult: RunResult{
					RunMode:        PreviewMode,
					Sources:        nil,
					Specifications: nil,
					Artifacts:      nil,
					StartedAt:      testDay,
					EndedAt:        testDay,
				},
				expectedError: assert.NoError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.given.specter()

			actualResult, err := s.Run(tt.when.context, tt.when.sourceLocations, tt.when.executionMode)
			if tt.then.expectedError != nil {
				tt.then.expectedError(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.then.expectedRunResult, actualResult)
		})
	}
}
