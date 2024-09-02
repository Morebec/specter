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
	. "github.com/morebec/specter/pkg/specterutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// MockDependencyProvider is a mock implementation of DependencyProvider for testing.
type MockDependencyProvider struct {
	supportFunc func(s specter.Specification) bool
	provideFunc func(s specter.Specification) []specter.SpecificationName
}

func (m *MockDependencyProvider) Supports(s specter.Specification) bool {
	return m.supportFunc(s)
}

func (m *MockDependencyProvider) Provide(s specter.Specification) []specter.SpecificationName {
	return m.provideFunc(s)
}

func TestDependencyResolutionProcessor_Process(t *testing.T) {
	type args struct {
		specifications []specter.Specification
		providers      []DependencyProvider
	}
	spec1 := NewGenericSpecification("spec1", "type", specter.Source{})
	spec2 := NewGenericSpecification("spec2", "type", specter.Source{})
	tests := []struct {
		name          string
		given         args
		then          ResolvedDependencies
		expectedError error
	}{
		{
			name: "GIVEN no providers THEN returns nil",
			given: args{
				providers:      nil,
				specifications: nil,
			},
			then:          nil,
			expectedError: nil,
		},
		{
			name: "GIVEN a simple acyclic graph and multiple providers WHEN resolved THEN returns resolved dependencies",
			given: args{
				providers: []DependencyProvider{
					&MockDependencyProvider{
						supportFunc: func(s specter.Specification) bool {
							return false
						},
						provideFunc: func(s specter.Specification) []specter.SpecificationName {
							return nil
						},
					},
					&MockDependencyProvider{
						supportFunc: func(s specter.Specification) bool {
							return s.Type() == "type"
						},
						provideFunc: func(s specter.Specification) []specter.SpecificationName {
							if s.Name() == "spec1" {
								return []specter.SpecificationName{"spec2"}
							}
							return nil
						},
					},
				},
				specifications: specter.SpecificationGroup{
					spec1,
					spec2,
				},
			},
			then: ResolvedDependencies{
				spec2, // topological sort
				spec1,
			},
			expectedError: nil,
		},
		{
			name: "GIVEN circular dependencies WHEN resolved THEN returns an error",
			given: args{
				providers: []DependencyProvider{
					&MockDependencyProvider{
						supportFunc: func(s specter.Specification) bool {
							return s.Type() == "type"
						},
						provideFunc: func(s specter.Specification) []specter.SpecificationName {
							if s.Name() == "spec1" {
								return []specter.SpecificationName{"spec2"}
							} else if s.Name() == "spec2" {
								return []specter.SpecificationName{"spec1"}
							}
							return nil
						},
					},
				},
				specifications: specter.SpecificationGroup{
					spec1,
					spec2,
				},
			},
			then:          nil,
			expectedError: errors.NewWithMessage(errors.InternalErrorCode, "circular dependencies found"),
		},
		{
			name: "GIVEN unresolvable dependencies THEN returns an error",
			given: args{
				providers: []DependencyProvider{
					&MockDependencyProvider{
						supportFunc: func(s specter.Specification) bool {
							return s.Type() == "type"
						},
						provideFunc: func(s specter.Specification) []specter.SpecificationName {
							return []specter.SpecificationName{"spec3"}
						},
					},
				},
				specifications: specter.SpecificationGroup{
					spec1,
					spec2, // spec2 is not provided
				},
			},
			then:          nil,
			expectedError: errors.NewWithMessage(errors.InternalErrorCode, "depends on an unresolved type \"spec3\""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewDependencyResolutionProcessor(tt.given.providers...)

			ctx := specter.ProcessingContext{
				Specifications: tt.given.specifications,
				Logger: specter.NewDefaultLogger(specter.DefaultLoggerConfig{
					DisableColors: true,
					Writer:        os.Stdout,
				}),
			}

			var err error
			ctx.Artifacts, err = processor.Process(ctx)
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedError.Error())
				return
			}

			artifact := ctx.Artifact(ResolvedDependenciesArtifactID)
			graph := artifact.(ResolvedDependencies)

			require.NoError(t, err)
			require.Equal(t, tt.then, graph)
		})
	}
}

func TestGetResolvedDependenciesFromContext(t *testing.T) {
	tests := []struct {
		name  string
		given specter.ProcessingContext
		want  ResolvedDependencies
	}{
		{
			name: "GIVEN a context with resolved dependencies THEN return resolved dependencies",
			given: specter.ProcessingContext{
				Artifacts: []specter.Artifact{
					ResolvedDependencies{
						NewGenericSpecification("name", "type", specter.Source{}),
					},
				},
			},
			want: ResolvedDependencies{
				NewGenericSpecification("name", "type", specter.Source{}),
			},
		},
		{
			name: "GIVEN a context with resolved dependencies with wrong type THEN return nil",
			given: specter.ProcessingContext{
				Artifacts: []specter.Artifact{
					NewArtifactStub(ResolvedDependenciesArtifactID),
				},
			},
			want: nil,
		},
		{
			name:  "GIVEN a context without resolved dependencies THEN return nil",
			given: specter.ProcessingContext{},
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := GetResolvedDependenciesFromContext(tt.given)
			assert.Equal(t, tt.want, deps)
		})
	}
}

func TestDependencyResolutionProcessor_Name(t *testing.T) {
	p := DependencyResolutionProcessor{}
	assert.NotEqual(t, "", p.Name())
}

type hasDependencySpec struct {
	source       specter.Source
	dependencies []specter.SpecificationName
}

func (h *hasDependencySpec) Name() specter.SpecificationName {
	return "spec"
}

func (h *hasDependencySpec) Type() specter.SpecificationType {
	return "spec"
}

func (h *hasDependencySpec) Description() string {
	return "description"
}

func (h *hasDependencySpec) Source() specter.Source {
	return h.source
}

func (h *hasDependencySpec) SetSource(s specter.Source) {
	h.source = s
}

func (h *hasDependencySpec) Dependencies() []specter.SpecificationName {
	return h.dependencies
}

func TestHasDependenciesProvider_Supports(t *testing.T) {
	tests := []struct {
		name  string
		given specter.Specification
		then  bool
	}{
		{
			name:  "GIVEN specification not implementing HasDependencies THEN return false",
			given: &GenericSpecification{},
			then:  false,
		},
		{
			name:  "GIVEN specification implementing HasDependencies THEN return false",
			given: &hasDependencySpec{},
			then:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HasDependenciesProvider{}
			assert.Equal(t, tt.then, h.Supports(tt.given))
		})
	}
}

func TestHasDependenciesProvider_Provide(t *testing.T) {
	tests := []struct {
		name  string
		given specter.Specification
		then  []specter.SpecificationName
	}{
		{
			name:  "GIVEN specification not implementing HasDependencies THEN return nil",
			given: &GenericSpecification{},
			then:  nil,
		},
		{
			name:  "GIVEN specification implementing HasDependencies THEN return dependencies",
			given: &hasDependencySpec{dependencies: []specter.SpecificationName{"spec1"}},
			then:  []specter.SpecificationName{"spec1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HasDependenciesProvider{}
			assert.Equal(t, tt.then, h.Provide(tt.given))
		})
	}
}
