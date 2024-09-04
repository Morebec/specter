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
	"github.com/zclconf/go-cty/cty"
	"testing"
)

func TestUnitsDescriptionsMustStartWithACapitalLetter(t *testing.T) {
	tests := []struct {
		name  string
		given specter.UnitGroup
		then  specterutils.LinterResultSet
	}{
		{
			name: "GIVEN unit starting with an upper case letter THEN return empty result set",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitName: "test",
					Attributes: []specterutils.GenericUnitAttribute{
						{
							Name:  "description",
							Value: specterutils.GenericValue{Value: cty.StringVal("It starts with UPPERCASE")},
						},
					},
				},
			},
		},
		{
			name: "GIVEN unit starting with lower case letter THEN return error",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitName: "test",
					Attributes: []specterutils.GenericUnitAttribute{
						{
							Name:  "description",
							Value: specterutils.GenericValue{Value: cty.StringVal("it starts with lowercase")},
						},
					},
				},
			},
			then: specterutils.LinterResultSet{
				{
					Severity: specterutils.ErrorSeverity,
					Message:  "the description of unit \"test\" at location \"\" does not start with a capital letter",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := specterutils.UnitsDescriptionsMustStartWithACapitalLetter(specterutils.ErrorSeverity)
			result := linter.Lint(tt.given)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestUnitsDescriptionsMustEndWithPeriod(t *testing.T) {
	tests := []struct {
		name  string
		given specter.UnitGroup
		then  specterutils.LinterResultSet
	}{
		{
			name: "GIVEN unit ending with period THEN return empty result set",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitName: "test",
					Attributes: []specterutils.GenericUnitAttribute{
						{
							Name:  "description",
							Value: specterutils.GenericValue{Value: cty.StringVal("it ends with period.")},
						},
					},
				},
			},
		},
		{
			name: "GIVEN unit not ending with period THEN return error",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitName: "test",
					Attributes: []specterutils.GenericUnitAttribute{
						{
							Name:  "description",
							Value: specterutils.GenericValue{Value: cty.StringVal("it starts with lowercase")},
						},
					},
				},
			},
			then: specterutils.LinterResultSet{
				{
					Severity: specterutils.ErrorSeverity,
					Message:  "the description of unit \"test\" at location \"\" does not end with a period",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := specterutils.UnitsDescriptionsMustEndWithPeriod(specterutils.ErrorSeverity)
			result := linter.Lint(tt.given)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestUnitsMustHaveDescriptionAttribute(t *testing.T) {
	tests := []struct {
		name  string
		given specter.UnitGroup
		then  specterutils.LinterResultSet
	}{
		{
			name: "GIVEN unit with a description THEN return empty result set",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitName: "test",
					Attributes: []specterutils.GenericUnitAttribute{
						{
							Name:  "description",
							Value: specterutils.GenericValue{Value: cty.StringVal("I have a description")},
						},
					},
				},
			},
		},
		{
			name: "GIVEN unit with no description ",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitName: "test",
				},
			},
			then: specterutils.LinterResultSet{
				{
					Severity: specterutils.ErrorSeverity,
					Message:  "unit \"test\" at location \"\" does not have a description",
				},
			},
		},
		{
			name: "GIVEN unit with an empty description THEN return error",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitName: "test",
					Attributes: []specterutils.GenericUnitAttribute{
						{
							Name:  "description",
							Value: specterutils.GenericValue{Value: cty.StringVal("")},
						},
					},
				},
			},
			then: specterutils.LinterResultSet{
				{
					Severity: specterutils.ErrorSeverity,
					Message:  "unit \"test\" at location \"\" does not have a description",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := specterutils.UnitsMustHaveDescriptionAttribute(specterutils.ErrorSeverity)
			result := linter.Lint(tt.given)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestUnitsMustHaveUniqueNames(t *testing.T) {
	tests := []struct {
		name  string
		given specter.UnitGroup
		then  specterutils.LinterResultSet
	}{
		{
			name: "GIVEN units with unique names THEN return empty result set",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitName: "test",
				},
				&specterutils.GenericUnit{
					UnitName: "test2",
				},
			},
		},
		{
			name: "GIVEN units with non-unique names THEN return error",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitName: "test",
				},
				&specterutils.GenericUnit{
					UnitName: "test",
				},
			},
			then: specterutils.LinterResultSet{
				{
					Severity: specterutils.ErrorSeverity,
					Message:  "duplicate unit name detected for \"test\" in the following file(s): ",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := specterutils.UnitsMustHaveUniqueNames(specterutils.ErrorSeverity)
			result := linter.Lint(tt.given)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestUnitMustNotHaveUndefinedNames(t *testing.T) {
	tests := []struct {
		name  string
		given specter.UnitGroup
		then  specterutils.LinterResultSet
	}{
		{
			name: "GIVEN unit with a name THEN return empty result set",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitName: "test",
				},
			},
		},
		{
			name: "GIVEN unit with no name THEN return error ",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitName: "",
				},
			},
			then: specterutils.LinterResultSet{
				{
					Severity: specterutils.ErrorSeverity,
					Message:  "unit at \"\" has an undefined name",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := specterutils.UnitMustNotHaveUndefinedNames(specterutils.ErrorSeverity)
			result := linter.Lint(tt.given)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestCompositeUnitLinter(t *testing.T) {
	type args struct {
		linters []specterutils.UnitLinter
		units   specter.UnitGroup
	}
	tests := []struct {
		name  string
		given args
		then  specterutils.LinterResultSet
	}{
		{
			name: "GIVEN valid units THEN return empty result set",
			given: args{
				linters: []specterutils.UnitLinter{
					specterutils.UnitMustNotHaveUndefinedNames(specterutils.ErrorSeverity),
					specterutils.UnitsDescriptionsMustStartWithACapitalLetter(specterutils.ErrorSeverity),
					specterutils.UnitsDescriptionsMustEndWithPeriod(specterutils.ErrorSeverity),
				},
				units: specter.UnitGroup{
					&specterutils.GenericUnit{
						UnitName: "test",
						Attributes: []specterutils.GenericUnitAttribute{
							{
								Name:  "description",
								Value: specterutils.GenericValue{Value: cty.StringVal("This is a Description.")},
							},
						},
					},
				},
			},
		},
		{
			name: "GIVEN invalid units THEN return empty result set",
			given: args{
				linters: []specterutils.UnitLinter{
					specterutils.UnitMustNotHaveUndefinedNames(specterutils.ErrorSeverity),
					specterutils.UnitsDescriptionsMustStartWithACapitalLetter(specterutils.ErrorSeverity),
					specterutils.UnitsDescriptionsMustEndWithPeriod(specterutils.ErrorSeverity),
				},
				units: specter.UnitGroup{
					&specterutils.GenericUnit{
						UnitName: "",
					},
				},
			},
			then: specterutils.LinterResultSet{
				{
					Severity: specterutils.ErrorSeverity,
					Message:  "unit at \"\" has an undefined name",
				},
				{
					Severity: specterutils.ErrorSeverity,
					Message:  "the description of unit \"\" at location \"\" does not start with a capital letter",
				},

				{
					Severity: specterutils.ErrorSeverity,
					Message:  "the description of unit \"\" at location \"\" does not end with a period",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := specterutils.CompositeUnitLinter(tt.given.linters...)
			result := linter.Lint(tt.given.units)
			require.Equal(t, tt.then, result)
		})
	}
}

func TestLinterResultSet_HasWarnings(t *testing.T) {
	tests := []struct {
		name  string
		given specterutils.LinterResultSet
		then  bool
	}{
		{
			name: "GIVEN result set with some warnings THEN return true",
			given: specterutils.LinterResultSet{
				specterutils.LinterResult{
					Severity: specterutils.WarningSeverity,
					Message:  "a warning",
				},
			},
			then: true,
		},
		{
			name: "GIVEN result set with no warnings THEN return false",
			given: specterutils.LinterResultSet{
				specterutils.LinterResult{
					Severity: specterutils.ErrorSeverity,
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
		given specterutils.LinterResultSet
		then  bool
	}{
		{
			name: "GIVEN result set with some error THEN return true",
			given: specterutils.LinterResultSet{
				specterutils.LinterResult{
					Severity: specterutils.ErrorSeverity,
					Message:  assert.AnError.Error(),
				},
			},
			then: true,
		},
		{
			name: "GIVEN result set with no error THEN return false",
			given: specterutils.LinterResultSet{
				specterutils.LinterResult{
					Severity: specterutils.WarningSeverity,
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
		given specterutils.LinterResultSet
		then  specterutils.LinterResultSet
	}{
		{
			name: "GIVEN a result set with some warnings THEN return warnings",
			given: specterutils.LinterResultSet{
				specterutils.LinterResult{
					Severity: specterutils.WarningSeverity,
					Message:  "a warning",
				},
				specterutils.LinterResult{
					Severity: specterutils.ErrorSeverity,
					Message:  assert.AnError.Error(),
				},
			},
			then: specterutils.LinterResultSet{
				specterutils.LinterResult{
					Severity: specterutils.WarningSeverity,
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
		given specterutils.LinterResultSet
		then  errors.Group
	}{
		{
			name: "GIVEN a result set with some errors THEN return errors",
			given: specterutils.LinterResultSet{
				specterutils.LinterResult{
					Severity: specterutils.WarningSeverity,
					Message:  "a warning",
				},
				specterutils.LinterResult{
					Severity: specterutils.ErrorSeverity,
					Message:  assert.AnError.Error(),
				},
			},
			then: errors.NewGroup(specterutils.LintingErrorCode, errors.NewWithMessage(specterutils.LintingErrorCode, assert.AnError.Error())),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.given.Errors()
			require.Equal(t, specterutils.LintingErrorCode, errs.Code())
			assert.Equalf(t, tt.then, errs, "Warnings()")
		})
	}
}

func TestLintingProcessor_Name(t *testing.T) {
	lp := specterutils.LintingProcessor{}
	require.NotEqual(t, "", lp.Name())
}

func TestLintingProcessor_Process(t *testing.T) {
	type args struct {
		linters []specterutils.UnitLinter
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
				specterutils.LinterResultSet(nil),
			},
			expectedError: nil,
		},
		{
			name: "GIVEN a processing context with units that raise warnings THEN return a processing artifact with the result set",
			given: args{
				linters: []specterutils.UnitLinter{
					specterutils.UnitLinterFunc(func(units specter.UnitGroup) specterutils.LinterResultSet {
						return specterutils.LinterResultSet{{Severity: specterutils.WarningSeverity, Message: "a warning"}}
					}),
				},
				ctx: specter.ProcessingContext{
					Units:  []specter.Unit{specterutils.NewGenericUnit("unit", "spec_type", specter.Source{})},
					Logger: specter.NewDefaultLogger(specter.DefaultLoggerConfig{}),
				},
			},
			then: []specter.Artifact{
				specterutils.LinterResultSet{{Severity: specterutils.WarningSeverity, Message: "a warning"}},
			},
		},
		{
			name: "GIVEN a processing context that will raise errors THEN return errors",
			given: args{
				linters: []specterutils.UnitLinter{
					specterutils.UnitLinterFunc(func(units specter.UnitGroup) specterutils.LinterResultSet {
						return specterutils.LinterResultSet{{Severity: specterutils.ErrorSeverity, Message: assert.AnError.Error()}}
					}),
				},
				ctx: specter.ProcessingContext{
					Units:  []specter.Unit{specterutils.NewGenericUnit("unit", "spec_type", specter.Source{})},
					Logger: specter.NewDefaultLogger(specter.DefaultLoggerConfig{}),
				},
			},
			then: []specter.Artifact{
				specterutils.LinterResultSet{{Severity: specterutils.ErrorSeverity, Message: assert.AnError.Error()}},
			},
			expectedError: assert.AnError,
		},
		{
			name: "GIVEN a processing context that will raise both errors and warnings THEN return errors and warnings",
			given: args{
				linters: []specterutils.UnitLinter{
					specterutils.UnitLinterFunc(func(units specter.UnitGroup) specterutils.LinterResultSet {
						return specterutils.LinterResultSet{
							{
								Severity: specterutils.ErrorSeverity, Message: assert.AnError.Error(),
							},
							{
								Severity: specterutils.WarningSeverity, Message: "a warning",
							},
						}
					}),
				},
				ctx: specter.ProcessingContext{
					Units:  []specter.Unit{specterutils.NewGenericUnit("unit", "spec_type", specter.Source{})},
					Logger: specter.NewDefaultLogger(specter.DefaultLoggerConfig{}),
				},
			},
			then: []specter.Artifact{
				specterutils.LinterResultSet{
					{
						Severity: specterutils.ErrorSeverity, Message: assert.AnError.Error(),
					},
					{
						Severity: specterutils.WarningSeverity, Message: "a warning",
					},
				},
			},
			expectedError: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := specterutils.NewLintingProcessor(tt.given.linters...)
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
		then  specterutils.LinterResultSet
	}{
		{
			name: "GIVEN context with linting results THEN return linting results",
			given: specter.ProcessingContext{
				Artifacts: []specter.Artifact{
					specterutils.LinterResultSet{{Severity: specterutils.WarningSeverity, Message: "a warning"}},
				},
			},
			then: specterutils.LinterResultSet{{Severity: specterutils.WarningSeverity, Message: "a warning"}},
		},
		{
			name:  "GIVEN context with not linting results THEN return empty linting results",
			given: specter.ProcessingContext{},
			then:  specterutils.LinterResultSet(nil),
		},
		{
			name: "GIVEN a context with wrong value for artifact name THEN return nil",
			given: specter.ProcessingContext{
				Artifacts: []specter.Artifact{
					testutils.NewArtifactStub(specterutils.LinterResultArtifactID),
				},
			},
			then: specterutils.LinterResultSet(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := specterutils.GetLintingResultsFromContext(tt.given)
			assert.Equal(t, tt.then, got)
		})
	}
}
