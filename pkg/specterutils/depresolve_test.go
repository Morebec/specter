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

package specterutils_test

import (
	"github.com/morebec/go-errors/errors"
	"github.com/morebec/specter/pkg/specter"
	"github.com/morebec/specter/pkg/specterutils"
	"github.com/morebec/specter/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// MockDependencyProvider is a mock implementation of DependencyProvider for testing.
type MockDependencyProvider struct {
	supportFunc func(specter.Unit) bool
	provideFunc func(specter.Unit) []specter.UnitID
}

func (m *MockDependencyProvider) Supports(u specter.Unit) bool {
	return m.supportFunc(u)
}

func (m *MockDependencyProvider) Provide(u specter.Unit) []specter.UnitID {
	return m.provideFunc(u)
}

func TestDependencyResolutionProcessor_Process(t *testing.T) {
	type args struct {
		units     []specter.Unit
		providers []specterutils.DependencyProvider
	}
	unit1 := specterutils.NewGenericUnit("unit1", "type", specter.Source{})
	unit2 := specterutils.NewGenericUnit("unit2", "type", specter.Source{})
	tests := []struct {
		name          string
		given         args
		then          specterutils.ResolvedDependencies
		expectedError error
	}{
		{
			name: "GIVEN no providers THEN returns nil",
			given: args{
				providers: nil,
				units:     nil,
			},
			then:          nil,
			expectedError: nil,
		},
		{
			name: "GIVEN a simple acyclic graph and multiple providers WHEN resolved THEN returns resolved dependencies",
			given: args{
				providers: []specterutils.DependencyProvider{
					&MockDependencyProvider{
						supportFunc: func(_ specter.Unit) bool {
							return false
						},
						provideFunc: func(_ specter.Unit) []specter.UnitID {
							return nil
						},
					},
					&MockDependencyProvider{
						supportFunc: func(u specter.Unit) bool {
							return u.Kind() == "type"
						},
						provideFunc: func(u specter.Unit) []specter.UnitID {
							if u.ID() == "unit1" {
								return []specter.UnitID{"unit2"}
							}
							return nil
						},
					},
				},
				units: specter.UnitGroup{
					unit1,
					unit2,
				},
			},
			then: specterutils.ResolvedDependencies{
				unit2, // topological sort
				unit1,
			},
			expectedError: nil,
		},
		{
			name: "GIVEN circular dependencies WHEN resolved THEN returns an error",
			given: args{
				providers: []specterutils.DependencyProvider{
					&MockDependencyProvider{
						supportFunc: func(u specter.Unit) bool {
							return u.Kind() == "type"
						},
						provideFunc: func(u specter.Unit) []specter.UnitID {
							if u.ID() == "unit1" {
								return []specter.UnitID{"unit2"}
							} else if u.ID() == "unit2" {
								return []specter.UnitID{"unit1"}
							}
							return nil
						},
					},
				},
				units: specter.UnitGroup{
					unit1,
					unit2,
				},
			},
			then:          nil,
			expectedError: errors.NewWithMessage(errors.InternalErrorCode, "circular dependencies found"),
		},
		{
			name: "GIVEN unresolvable dependencies THEN returns an error",
			given: args{
				providers: []specterutils.DependencyProvider{
					&MockDependencyProvider{
						supportFunc: func(u specter.Unit) bool {
							return u.Kind() == "type"
						},
						provideFunc: func(u specter.Unit) []specter.UnitID {
							return []specter.UnitID{"spec3"}
						},
					},
				},
				units: specter.UnitGroup{
					unit1,
					unit2, // unit2 is not provided
				},
			},
			then:          nil,
			expectedError: errors.NewWithMessage(specterutils.DependencyResolutionFailed, "depends on an unresolved kind \"spec3\""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := specterutils.NewDependencyResolutionProcessor(tt.given.providers...)

			ctx := specter.UnitProcessingContext{
				Units: tt.given.units,
			}

			var err error
			ctx.Artifacts, err = processor.Process(ctx)
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedError.Error())
				return
			}

			artifact := ctx.Artifact(specterutils.ResolvedDependenciesArtifactID)
			graph := artifact.(specterutils.ResolvedDependencies)

			require.NoError(t, err)
			require.Equal(t, tt.then, graph)
		})
	}
}

func TestGetResolvedDependenciesFromContext(t *testing.T) {
	tests := []struct {
		name  string
		given specter.UnitProcessingContext
		want  specterutils.ResolvedDependencies
	}{
		{
			name: "GIVEN a context with resolved dependencies THEN return resolved dependencies",
			given: specter.UnitProcessingContext{
				Artifacts: []specter.Artifact{
					specterutils.ResolvedDependencies{
						specterutils.NewGenericUnit("name", "type", specter.Source{}),
					},
				},
			},
			want: specterutils.ResolvedDependencies{
				specterutils.NewGenericUnit("name", "type", specter.Source{}),
			},
		},
		{
			name: "GIVEN a context with resolved dependencies with wrong type THEN return nil",
			given: specter.UnitProcessingContext{
				Artifacts: []specter.Artifact{
					testutils.NewArtifactStub(specterutils.ResolvedDependenciesArtifactID),
				},
			},
			want: nil,
		},
		{
			name:  "GIVEN a context without resolved dependencies THEN return nil",
			given: specter.UnitProcessingContext{},
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := specterutils.GetResolvedDependenciesFromContext(tt.given)
			assert.Equal(t, tt.want, deps)
		})
	}
}

func TestDependencyResolutionProcessor_Name(t *testing.T) {
	p := specterutils.DependencyResolutionProcessor{}
	assert.NotEqual(t, "", p.Name())
}

type hasDependencyUnit struct {
	source       specter.Source
	dependencies []specter.UnitID
}

func (h *hasDependencyUnit) ID() specter.UnitID {
	return "unit"
}

func (h *hasDependencyUnit) Kind() specter.UnitKind {
	return "unit"
}

func (h *hasDependencyUnit) Description() string {
	return "description"
}

func (h *hasDependencyUnit) Source() specter.Source {
	return h.source
}

func (h *hasDependencyUnit) Dependencies() []specter.UnitID {
	return h.dependencies
}

func TestHasDependenciesProvider_Supports(t *testing.T) {
	tests := []struct {
		name  string
		given specter.Unit
		then  bool
	}{
		{
			name:  "GIVEN unit not implementing HasDependencies THEN return false",
			given: &specterutils.GenericUnit{},
			then:  false,
		},
		{
			name:  "GIVEN unit implementing HasDependencies THEN return false",
			given: &hasDependencyUnit{},
			then:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := specterutils.HasDependenciesProvider{}
			assert.Equal(t, tt.then, h.Supports(tt.given))
		})
	}
}

func TestHasDependenciesProvider_Provide(t *testing.T) {
	tests := []struct {
		name  string
		given specter.Unit
		then  []specter.UnitID
	}{
		{
			name:  "GIVEN unit not implementing HasDependencies THEN return nil",
			given: &specterutils.GenericUnit{},
			then:  nil,
		},
		{
			name:  "GIVEN unit implementing HasDependencies THEN return dependencies",
			given: &hasDependencyUnit{dependencies: []specter.UnitID{"unit1"}},
			then:  []specter.UnitID{"unit1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := specterutils.HasDependenciesProvider{}
			assert.Equal(t, tt.then, h.Provide(tt.given))
		})
	}
}
