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

func TestUnitsIDsMustBeUnique(t *testing.T) {
	tests := []struct {
		name  string
		given specter.UnitGroup
		then  specterutils.LinterResultSet
	}{
		{
			name: "GIVEN units with unique names THEN return empty result set",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitID: "test",
				},
				&specterutils.GenericUnit{
					UnitID: "test2",
				},
			},
		},
		{
			name: "GIVEN units with non-unique IDs THEN return error",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitID: "test",
				},
				&specterutils.GenericUnit{
					UnitID: "test",
				},
			},
			then: specterutils.LinterResultSet{
				{
					Severity: specterutils.ErrorSeverity,
					Message:  "duplicate unit ID detected \"test\" in the following file(s): ",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := specterutils.UnitsIDsMustBeUnique(specterutils.ErrorSeverity)
			result := linter.Lint(tt.given)
			require.Equal(t, tt.then, result)
		})
	}
}

func UnitsMustHaveIDs(t *testing.T) {
	tests := []struct {
		name  string
		given specter.UnitGroup
		then  specterutils.LinterResultSet
	}{
		{
			name: "GIVEN unit with a name THEN return empty result set",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitID: "test",
				},
			},
		},
		{
			name: "GIVEN unit with no name THEN return error ",
			given: specter.UnitGroup{
				&specterutils.GenericUnit{
					UnitID: "",
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
			linter := specterutils.UnitsMustHaveIDs(specterutils.ErrorSeverity)
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
					specterutils.UnitsMustHaveIDs(specterutils.ErrorSeverity),
					specterutils.UnitsIDsMustBeUnique(specterutils.ErrorSeverity),
				},
				units: specter.UnitGroup{
					&specterutils.GenericUnit{
						UnitID: "test",
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
					specterutils.UnitsMustHaveIDs(specterutils.ErrorSeverity),
				},
				units: specter.UnitGroup{
					&specterutils.GenericUnit{
						UnitID: "", // invalid because of ID
					},
				},
			},
			then: specterutils.LinterResultSet{
				{
					Severity: specterutils.ErrorSeverity,
					Message:  "a unit of kind \"\" has no ID at \"\"",
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
		ctx     specter.UnitProcessingContext
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
				ctx:     specter.UnitProcessingContext{},
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
				ctx: specter.UnitProcessingContext{
					Units: []specter.Unit{specterutils.NewGenericUnit("unit", "spec_type", specter.Source{})},
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
				ctx: specter.UnitProcessingContext{
					Units: []specter.Unit{specterutils.NewGenericUnit("unit", "spec_type", specter.Source{})},
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
				ctx: specter.UnitProcessingContext{
					Units: []specter.Unit{specterutils.NewGenericUnit("unit", "spec_type", specter.Source{})},
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
		given specter.UnitProcessingContext
		then  specterutils.LinterResultSet
	}{
		{
			name: "GIVEN context with linting results THEN return linting results",
			given: specter.UnitProcessingContext{
				Artifacts: []specter.Artifact{
					specterutils.LinterResultSet{{Severity: specterutils.WarningSeverity, Message: "a warning"}},
				},
			},
			then: specterutils.LinterResultSet{{Severity: specterutils.WarningSeverity, Message: "a warning"}},
		},
		{
			name:  "GIVEN context with not linting results THEN return empty linting results",
			given: specter.UnitProcessingContext{},
			then:  specterutils.LinterResultSet(nil),
		},
		{
			name: "GIVEN a context with wrong value for artifact name THEN return nil",
			given: specter.UnitProcessingContext{
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
