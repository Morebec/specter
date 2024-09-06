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
	"github.com/morebec/specter/pkg/specter"
	"github.com/morebec/specter/pkg/specterutils"
	"github.com/stretchr/testify/require"
	"testing"
)

// Mock implementations for the Unit and HasVersion interfaces

var _ specter.Unit = (*mockUnit)(nil)

type mockUnit struct {
	name        specter.UnitID
	description string
	source      specter.Source
	version     specterutils.UnitVersion
	kind        specter.UnitKind
}

func (m *mockUnit) ID() specter.UnitID {
	return m.name
}

func (m *mockUnit) Kind() specter.UnitKind {
	return m.kind
}

func (m *mockUnit) Description() string {
	return m.description
}

func (m *mockUnit) Source() specter.Source {
	return m.source
}

func (m *mockUnit) Version() specterutils.UnitVersion {
	return m.version
}

func TestHasVersionMustHaveAVersionLinter(t *testing.T) {
	tests := []struct {
		name            string
		when            specter.UnitGroup
		expectedResults specterutils.LinterResultSet
		givenSeverity   specterutils.LinterResultSeverity
	}{
		{
			name: "WHEN some unit does not implement HasVersion, THEN ignore said unit",
			when: specter.UnitGroup{
				&mockUnit{name: "spec1", version: "v1"},
				specterutils.NewGenericUnit("not-versioned", "unit", specter.Source{}),
			},
			givenSeverity:   specterutils.WarningSeverity,
			expectedResults: specterutils.LinterResultSet(nil),
		},
		{
			name: "WHEN all units have a version THEN return no warnings or errors",
			when: specter.UnitGroup{
				&mockUnit{name: "spec1", version: "v1"},
				&mockUnit{name: "spec2", version: "v2"},
			},
			givenSeverity:   specterutils.WarningSeverity,
			expectedResults: specterutils.LinterResultSet(nil),
		},
		{
			name: "WHEN one unit is missing a version and severity is Warning THEN return a warning",
			when: specter.UnitGroup{
				&mockUnit{name: "spec1", version: "v1"},
				&mockUnit{name: "spec2", version: ""},
			},
			givenSeverity: specterutils.WarningSeverity,
			expectedResults: specterutils.LinterResultSet{
				{
					Severity: specterutils.WarningSeverity,
					Message:  `unit "spec2" at "" should have a version`,
				},
			},
		},
		{
			name: "WHEN one unit is missing a version and severity is error THEN return an error",
			when: specter.UnitGroup{
				&mockUnit{name: "spec1", version: "v1"},
				&mockUnit{name: "spec2", version: ""},
			},
			givenSeverity: specterutils.ErrorSeverity,
			expectedResults: specterutils.LinterResultSet{
				{
					Severity: specterutils.ErrorSeverity,
					Message:  `unit "spec2" at "" should have a version`,
				},
			},
		},
		{
			name: "multiple units are missing versions",
			when: specter.UnitGroup{
				&mockUnit{name: "spec1", version: ""},
				&mockUnit{name: "spec2", version: ""},
			},
			givenSeverity: specterutils.ErrorSeverity,
			expectedResults: specterutils.LinterResultSet{
				{
					Severity: specterutils.ErrorSeverity,
					Message:  `unit "spec1" at "" should have a version`,
				},
				{
					Severity: specterutils.ErrorSeverity,
					Message:  `unit "spec2" at "" should have a version`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := specterutils.HasVersionMustHaveAVersionLinter(tt.givenSeverity)
			results := linter.Lint(tt.when)
			require.Equal(t, tt.expectedResults, results)
		})
	}
}
