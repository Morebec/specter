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

package specterutils

import (
	"github.com/morebec/go-errors/errors"
	"github.com/morebec/specter/pkg/specter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"testing"
)

func TestSpecificationsDescriptionsMustStartWithACapitalLetter(t *testing.T) {
	tests := []struct {
		name  string
		given specter.SpecificationGroup
		then  LinterResultSet
	}{
		{
			name: "GIVEN specification starting with an upper case letter THEN return empty result set",
			given: specter.SpecificationGroup{
				&GenericSpecification{
					name: "test",
					Attributes: []GenericSpecAttribute{
						{
							Name:  "description",
							Value: GenericValue{cty.StringVal("It starts with UPPERCASE")},
						},
					},
				},
			},
		},
		{
			name: "GIVEN specification starting with lower case letter THEN return error",
			given: specter.SpecificationGroup{
				&GenericSpecification{
					name: "test",
					Attributes: []GenericSpecAttribute{
						{
							Name:  "description",
							Value: GenericValue{cty.StringVal("it starts with lowercase")},
						},
					},
				},
			},
			then: LinterResultSet{
				{
					Severity: ErrorSeverity,
					Message:  "the description of specification \"test\" at location \"\" does not start with a capital letter",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := SpecificationsDescriptionsMustStartWithACapitalLetter(ErrorSeverity)
			result := linter.Lint(tt.given)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestSpecificationsDescriptionsMustEndWithPeriod(t *testing.T) {
	tests := []struct {
		name  string
		given specter.SpecificationGroup
		then  LinterResultSet
	}{
		{
			name: "GIVEN specification ending with period THEN return empty result set",
			given: specter.SpecificationGroup{
				&GenericSpecification{
					name: "test",
					Attributes: []GenericSpecAttribute{
						{
							Name:  "description",
							Value: GenericValue{cty.StringVal("it ends with period.")},
						},
					},
				},
			},
		},
		{
			name: "GIVEN specification not ending with period THEN return error",
			given: specter.SpecificationGroup{
				&GenericSpecification{
					name: "test",
					Attributes: []GenericSpecAttribute{
						{
							Name:  "description",
							Value: GenericValue{cty.StringVal("it starts with lowercase")},
						},
					},
				},
			},
			then: LinterResultSet{
				{
					Severity: ErrorSeverity,
					Message:  "the description of specification \"test\" at location \"\" does not end with a period",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := SpecificationsDescriptionsMustEndWithPeriod(ErrorSeverity)
			result := linter.Lint(tt.given)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestSpecificationsMustHaveDescriptionAttribute(t *testing.T) {
	tests := []struct {
		name  string
		given specter.SpecificationGroup
		then  LinterResultSet
	}{
		{
			name: "GIVEN specification with a description THEN return empty result set",
			given: specter.SpecificationGroup{
				&GenericSpecification{
					name: "test",
					Attributes: []GenericSpecAttribute{
						{
							Name:  "description",
							Value: GenericValue{cty.StringVal("I have a description")},
						},
					},
				},
			},
		},
		{
			name: "GIVEN specification with no description ",
			given: specter.SpecificationGroup{
				&GenericSpecification{
					name: "test",
				},
			},
			then: LinterResultSet{
				{
					Severity: ErrorSeverity,
					Message:  "specification \"test\" at location \"\" does not have a description",
				},
			},
		},
		{
			name: "GIVEN specification with an empty description THEN return error",
			given: specter.SpecificationGroup{
				&GenericSpecification{
					name: "test",
					Attributes: []GenericSpecAttribute{
						{
							Name:  "description",
							Value: GenericValue{cty.StringVal("")},
						},
					},
				},
			},
			then: LinterResultSet{
				{
					Severity: ErrorSeverity,
					Message:  "specification \"test\" at location \"\" does not have a description",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := SpecificationsMustHaveDescriptionAttribute(ErrorSeverity)
			result := linter.Lint(tt.given)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestSpecificationsMustHaveUniqueNames(t *testing.T) {
	tests := []struct {
		name  string
		given specter.SpecificationGroup
		then  LinterResultSet
	}{
		{
			name: "GIVEN specifications with unique names THEN return empty result set",
			given: specter.SpecificationGroup{
				&GenericSpecification{
					name: "test",
				},
				&GenericSpecification{
					name: "test2",
				},
			},
		},
		{
			name: "GIVEN specifications with non-unique names THEN return error",
			given: specter.SpecificationGroup{
				&GenericSpecification{
					name: "test",
				},
				&GenericSpecification{
					name: "test",
				},
			},
			then: LinterResultSet{
				{
					Severity: ErrorSeverity,
					Message:  "duplicate specification name detected for \"test\" in the following file(s): ",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := SpecificationsMustHaveUniqueNames(ErrorSeverity)
			result := linter.Lint(tt.given)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestSpecificationMustNotHaveUndefinedNames(t *testing.T) {
	tests := []struct {
		name  string
		given specter.SpecificationGroup
		then  LinterResultSet
	}{
		{
			name: "GIVEN specification with a name THEN return empty result set",
			given: specter.SpecificationGroup{
				&GenericSpecification{
					name: "test",
				},
			},
		},
		{
			name: "GIVEN specification with no name THEN return error ",
			given: specter.SpecificationGroup{
				&GenericSpecification{
					name: "",
				},
			},
			then: LinterResultSet{
				{
					Severity: ErrorSeverity,
					Message:  "specification at \"\" has an undefined name",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := SpecificationMustNotHaveUndefinedNames(ErrorSeverity)
			result := linter.Lint(tt.given)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestCompositeSpecificationLinter(t *testing.T) {
	type args struct {
		linters        []SpecificationLinter
		specifications specter.SpecificationGroup
	}
	tests := []struct {
		name  string
		given args
		then  LinterResultSet
	}{
		{
			name: "GIVEN valid specifications THEN return empty result set",
			given: args{
				linters: []SpecificationLinter{
					SpecificationMustNotHaveUndefinedNames(ErrorSeverity),
					SpecificationsDescriptionsMustStartWithACapitalLetter(ErrorSeverity),
					SpecificationsDescriptionsMustEndWithPeriod(ErrorSeverity),
				},
				specifications: specter.SpecificationGroup{
					&GenericSpecification{
						name: "test",
						Attributes: []GenericSpecAttribute{
							{
								Name:  "description",
								Value: GenericValue{cty.StringVal("This is a Description.")},
							},
						},
					},
				},
			},
		},
		{
			name: "GIVEN invalid specifications THEN return empty result set",
			given: args{
				linters: []SpecificationLinter{
					SpecificationMustNotHaveUndefinedNames(ErrorSeverity),
					SpecificationsDescriptionsMustStartWithACapitalLetter(ErrorSeverity),
					SpecificationsDescriptionsMustEndWithPeriod(ErrorSeverity),
				},
				specifications: specter.SpecificationGroup{
					&GenericSpecification{
						name: "",
					},
				},
			},
			then: LinterResultSet{
				{
					Severity: ErrorSeverity,
					Message:  "specification at \"\" has an undefined name",
				},
				{
					Severity: ErrorSeverity,
					Message:  "the description of specification \"\" at location \"\" does not start with a capital letter",
				},

				{
					Severity: ErrorSeverity,
					Message:  "the description of specification \"\" at location \"\" does not end with a period",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := CompositeSpecificationLinter(tt.given.linters...)
			result := linter.Lint(tt.given.specifications)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestLinterResultSet_HasWarnings(t *testing.T) {
	tests := []struct {
		name  string
		given LinterResultSet
		then  bool
	}{
		{
			name: "GIVEN result set with some warnings THEN return true",
			given: LinterResultSet{
				LinterResult{
					Severity: WarningSeverity,
					Message:  "a warning",
				},
			},
			then: true,
		},
		{
			name: "GIVEN result set with no warnings THEN return false",
			given: LinterResultSet{
				LinterResult{
					Severity: ErrorSeverity,
					Message:  assert.AnError.Error(),
				},
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.then, tt.given.HasWarnings(), "HasWarnings()")
		})
	}
}

func TestLinterResultSet_HasErrors(t *testing.T) {
	tests := []struct {
		name  string
		given LinterResultSet
		then  bool
	}{
		{
			name: "GIVEN result set with some error THEN return true",
			given: LinterResultSet{
				LinterResult{
					Severity: ErrorSeverity,
					Message:  assert.AnError.Error(),
				},
			},
			then: true,
		},
		{
			name: "GIVEN result set with no error THEN return false",
			given: LinterResultSet{
				LinterResult{
					Severity: WarningSeverity,
					Message:  "a warning",
				},
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.then, tt.given.HasErrors(), "HasWarnings()")
		})
	}
}

func TestLinterResultSet_Warnings(t *testing.T) {
	tests := []struct {
		name  string
		given LinterResultSet
		then  LinterResultSet
	}{
		{
			name: "GIVEN a result set with some warnings THEN return warnings",
			given: LinterResultSet{
				LinterResult{
					Severity: WarningSeverity,
					Message:  "a warning",
				},
				LinterResult{
					Severity: ErrorSeverity,
					Message:  assert.AnError.Error(),
				},
			},
			then: LinterResultSet{
				LinterResult{
					Severity: WarningSeverity,
					Message:  "a warning",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.then, tt.given.Warnings(), "Warnings()")
		})
	}
}

func TestLinterResultSet_Errors(t *testing.T) {
	tests := []struct {
		name  string
		given LinterResultSet
		then  errors.Group
	}{
		{
			name: "GIVEN a result set with some errors THEN return errors",
			given: LinterResultSet{
				LinterResult{
					Severity: WarningSeverity,
					Message:  "a warning",
				},
				LinterResult{
					Severity: ErrorSeverity,
					Message:  assert.AnError.Error(),
				},
			},
			then: errors.NewGroup(LintingErrorCode, errors.NewWithMessage(LintingErrorCode, assert.AnError.Error())),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.given.Errors()
			require.Equal(t, LintingErrorCode, errs.Code())
			assert.Equalf(t, tt.then, errs, "Warnings()")
		})
	}
}

func TestLintingProcessor_Name(t *testing.T) {
	lp := LintingProcessor{}
	require.NotEqual(t, "", lp.Name())
}

func TestLintingProcessor_Process(t *testing.T) {
	type args struct {
		linters []SpecificationLinter
		ctx     specter.ProcessingContext
	}
	tests := []struct {
		name          string
		given         args
		then          []specter.Artifact
		expectedError error
	}{
		{
			name: "GIVEN an empty processing context",
			given: args{
				linters: nil,
				ctx: specter.ProcessingContext{
					Logger: specter.NewDefaultLogger(specter.DefaultLoggerConfig{}),
				},
			},
			then: []specter.Artifact{
				LinterResultSet(nil),
			},
			expectedError: nil,
		},
		{
			name: "GIVEN a processing context with specifications that raise warnings THEN return a processing artifact with the result set",
			given: args{
				linters: []SpecificationLinter{
					SpecificationLinterFunc(func(specifications specter.SpecificationGroup) LinterResultSet {
						return LinterResultSet{{Severity: WarningSeverity, Message: "a warning"}}
					}),
				},
				ctx: specter.ProcessingContext{
					Specifications: []specter.Specification{NewGenericSpecification("spec", "spec_type", specter.Source{})},
					Logger:         specter.NewDefaultLogger(specter.DefaultLoggerConfig{}),
				},
			},
			then: []specter.Artifact{
				LinterResultSet{{Severity: WarningSeverity, Message: "a warning"}},
			},
		},
		{
			name: "GIVEN a processing context that will raise errors THEN return errors",
			given: args{
				linters: []SpecificationLinter{
					SpecificationLinterFunc(func(specifications specter.SpecificationGroup) LinterResultSet {
						return LinterResultSet{{Severity: ErrorSeverity, Message: assert.AnError.Error()}}
					}),
				},
				ctx: specter.ProcessingContext{
					Specifications: []specter.Specification{NewGenericSpecification("spec", "spec_type", specter.Source{})},
					Logger:         specter.NewDefaultLogger(specter.DefaultLoggerConfig{}),
				},
			},
			then: []specter.Artifact{
				LinterResultSet{{Severity: ErrorSeverity, Message: assert.AnError.Error()}},
			},
			expectedError: assert.AnError,
		},
		{
			name: "GIVEN a processing context that will raise both errors and warnings THEN return errors and warnings",
			given: args{
				linters: []SpecificationLinter{
					SpecificationLinterFunc(func(specifications specter.SpecificationGroup) LinterResultSet {
						return LinterResultSet{
							{
								Severity: ErrorSeverity, Message: assert.AnError.Error(),
							},
							{
								Severity: WarningSeverity, Message: "a warning",
							},
						}
					}),
				},
				ctx: specter.ProcessingContext{
					Specifications: []specter.Specification{NewGenericSpecification("spec", "spec_type", specter.Source{})},
					Logger:         specter.NewDefaultLogger(specter.DefaultLoggerConfig{}),
				},
			},
			then: []specter.Artifact{
				LinterResultSet{
					{
						Severity: ErrorSeverity, Message: assert.AnError.Error(),
					},
					{
						Severity: WarningSeverity, Message: "a warning",
					},
				},
			},
			expectedError: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLintingProcessor(tt.given.linters...)
			got, err := l.Process(tt.given.ctx)

			if tt.expectedError != nil {
				require.Error(t, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.then, got)
		})
	}
}

func TestGetLintingResultsFromContext(t *testing.T) {
	tests := []struct {
		name  string
		given specter.ProcessingContext
		then  LinterResultSet
	}{
		{
			name: "GIVEN context with linting results THEN return linting results",
			given: specter.ProcessingContext{
				Artifacts: []specter.Artifact{
					LinterResultSet{{Severity: WarningSeverity, Message: "a warning"}},
				},
			},
			then: LinterResultSet{{Severity: WarningSeverity, Message: "a warning"}},
		},
		{
			name:  "GIVEN context with not linting results THEN return empty linting results",
			given: specter.ProcessingContext{},
			then:  LinterResultSet(nil),
		},
		{
			name: "GIVEN a context with wrong value for artifact name THEN return nil",
			given: specter.ProcessingContext{
				Artifacts: []specter.Artifact{
					NewArtifactStub(LinterResultArtifactID),
				},
			},
			then: LinterResultSet(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLintingResultsFromContext(tt.given)
			assert.Equal(t, tt.then, got)
		})
	}
}
