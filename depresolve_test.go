package specter

import (
	"github.com/morebec/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// MockDependencyProvider is a mock implementation of DependencyProvider for testing.
type MockDependencyProvider struct {
	supportFunc func(s Specification) bool
	provideFunc func(s Specification) []SpecificationName
}

func (m *MockDependencyProvider) Supports(s Specification) bool {
	return m.supportFunc(s)
}

func (m *MockDependencyProvider) Provide(s Specification) []SpecificationName {
	return m.provideFunc(s)
}

func TestDependencyResolutionProcessor_Process(t *testing.T) {
	type args struct {
		specifications []Specification
		providers      []DependencyProvider
	}
	tests := []struct {
		name          string
		given         args
		then          ResolvedDependencies
		expectedError error
	}{
		{
			name: "GIVEN no matching providers THEN returns resolved dependencies in the same order",
			given: args{
				providers: []DependencyProvider{
					&MockDependencyProvider{
						supportFunc: func(s Specification) bool {
							return false
						},
						provideFunc: func(s Specification) []SpecificationName {
							return nil
						},
					},
					&MockDependencyProvider{
						supportFunc: func(s Specification) bool {
							return false
						},
						provideFunc: func(s Specification) []SpecificationName {
							if s.Name() == "spec1" {
								return []SpecificationName{"spec2"}
							}
							return nil
						},
					},
				},
				specifications: SpecificationGroup{
					NewGenericSpecification("spec1", "type", Source{}),
					NewGenericSpecification("spec2", "type", Source{}),
				},
			},
			then: ResolvedDependencies{
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
			},
			expectedError: nil,
		},
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
						supportFunc: func(s Specification) bool {
							return false
						},
						provideFunc: func(s Specification) []SpecificationName {
							return nil
						},
					},
					&MockDependencyProvider{
						supportFunc: func(s Specification) bool {
							return s.Type() == "type"
						},
						provideFunc: func(s Specification) []SpecificationName {
							if s.Name() == "spec1" {
								return []SpecificationName{"spec2"}
							}
							return nil
						},
					},
				},
				specifications: SpecificationGroup{
					NewGenericSpecification("spec1", "type", Source{}),
					NewGenericSpecification("spec2", "type", Source{}),
				},
			},
			then: ResolvedDependencies{
				NewGenericSpecification("spec2", "type", Source{}), // topological sort
				NewGenericSpecification("spec1", "type", Source{}),
			},
			expectedError: nil,
		},
		{
			name: "GIVEN circular dependencies WHEN resolved THEN returns an error",
			given: args{
				providers: []DependencyProvider{
					&MockDependencyProvider{
						supportFunc: func(s Specification) bool {
							return s.Type() == "type"
						},
						provideFunc: func(s Specification) []SpecificationName {
							if s.Name() == "spec1" {
								return []SpecificationName{"spec2"}
							} else if s.Name() == "spec2" {
								return []SpecificationName{"spec1"}
							}
							return nil
						},
					},
				},
				specifications: SpecificationGroup{
					NewGenericSpecification("spec1", "type", Source{}),
					NewGenericSpecification("spec2", "type", Source{}),
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
						supportFunc: func(s Specification) bool {
							return s.Type() == "type"
						},
						provideFunc: func(s Specification) []SpecificationName {
							return []SpecificationName{"spec3"}
						},
					},
				},
				specifications: SpecificationGroup{
					NewGenericSpecification("spec1", "type", Source{}),
					NewGenericSpecification("spec2", "type", Source{}), // spec2 is not provided
				},
			},
			then:          nil,
			expectedError: errors.NewWithMessage(errors.InternalErrorCode, "specification with type \"spec1\" depends on an unresolved type \"spec3\""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewDependencyResolutionProcessor(tt.given.providers...)

			ctx := ProcessingContext{
				Specifications: tt.given.specifications,
				Logger: NewColoredOutputLogger(ColoredOutputLoggerConfig{
					EnableColors: true,
					Writer:       os.Stdout,
				}),
			}

			var err error
			ctx.Outputs, err = processor.Process(ctx)
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedError.Error())
				return
			}

			output := ctx.Output(DependencyGraphContextName).Value
			graph := output.(ResolvedDependencies)

			require.NoError(t, err)
			require.Equal(t, tt.then, graph)
		})
	}
}
