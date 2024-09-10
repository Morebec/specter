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
	"github.com/morebec/specter/pkg/specter"
	"github.com/morebec/specter/pkg/testutils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUnitPreprocessorAdapter_Process(t *testing.T) {
	type given struct {
		PreprocessFunc func(specter.PipelineContext, []specter.Unit) ([]specter.Unit, error)
	}
	type when struct {
		ctx   specter.PipelineContext
		units []specter.Unit
	}
	type then struct {
		units []specter.Unit
		err   require.ErrorAssertionFunc
	}
	tests := []struct {
		name  string
		given given
		when  when
		then  then
	}{
		{
			name: "GIVEN no PreprocessFunc WHEN units THEN return same units and no error",
			given: given{
				PreprocessFunc: nil,
			},
			when: when{
				units: []specter.Unit{
					testutils.NewUnitStub("id", "kind", specter.Source{}),
				},
			},
			then: then{
				units: []specter.Unit{
					testutils.NewUnitStub("id", "kind", specter.Source{}),
				},
				err: require.NoError,
			},
		},
		{
			name: "GIVEN PreprocessFunc returns specific_units WHEN units THEN return specific_units and no error",
			given: given{
				PreprocessFunc: func(specter.PipelineContext, []specter.Unit) ([]specter.Unit, error) {
					return []specter.Unit{
						testutils.NewUnitStub("id", "kind", specter.Source{}),
					}, nil
				},
			},
			when: when{
				units: nil,
			},
			then: then{
				units: []specter.Unit{
					testutils.NewUnitStub("id", "kind", specter.Source{}),
				},
				err: require.NoError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := specter.UnitPreprocessorFunc("preprocessor", tt.given.PreprocessFunc)
			got, err := p.Preprocess(tt.when.ctx, tt.when.units)
			tt.then.err(t, err)
			require.Equal(t, tt.then.units, got)
		})
	}
}
