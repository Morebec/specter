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
	"github.com/morebec/specter/pkg/specter"
	"github.com/stretchr/testify/require"
	"testing"
)

// Mock implementations for the Specification and HasVersion interfaces

var _ specter.Specification = (*mockSpecification)(nil)

type mockSpecification struct {
	name        specter.SpecificationName
	description string
	source      specter.Source
	version     SpecificationVersion
	typeName    specter.SpecificationType
}

func (m *mockSpecification) Name() specter.SpecificationName {
	return m.name
}

func (m *mockSpecification) Type() specter.SpecificationType {
	return m.typeName
}

func (m *mockSpecification) Description() string {
	return m.description
}

func (m *mockSpecification) SetSource(s specter.Source) {
	m.source = s
}

func (m *mockSpecification) Source() specter.Source {
	return m.source
}

func (m *mockSpecification) Version() SpecificationVersion {
	return m.version
}

func TestHasVersionMustHaveAVersionLinter(t *testing.T) {
	tests := []struct {
		name            string
		when            specter.SpecificationGroup
		expectedResults LinterResultSet
		givenSeverity   LinterResultSeverity
	}{
		{
			name: "WHEN some specification does not implement HasVersion, THEN ignore said specification",
			when: specter.SpecificationGroup{
				&mockSpecification{name: "spec1", version: "v1"},
				NewGenericSpecification("not-versioned", "spec", specter.Source{}),
			},
			givenSeverity:   WarningSeverity,
			expectedResults: LinterResultSet(nil),
		},
		{
			name: "WHEN all specifications have a version THEN return no warnings or errors",
			when: specter.SpecificationGroup{
				&mockSpecification{name: "spec1", version: "v1"},
				&mockSpecification{name: "spec2", version: "v2"},
			},
			givenSeverity:   WarningSeverity,
			expectedResults: LinterResultSet(nil),
		},
		{
			name: "WHEN one specification is missing a version and severity is Warning THEN return a warning",
			when: specter.SpecificationGroup{
				&mockSpecification{name: "spec1", version: "v1"},
				&mockSpecification{name: "spec2", version: ""},
			},
			givenSeverity: WarningSeverity,
			expectedResults: LinterResultSet{
				{
					Severity: WarningSeverity,
					Message:  `specification "spec2" at "" should have a version`,
				},
			},
		},
		{
			name: "WHEN one specification is missing a version and severity is error THEN return an error",
			when: specter.SpecificationGroup{
				&mockSpecification{name: "spec1", version: "v1"},
				&mockSpecification{name: "spec2", version: ""},
			},
			givenSeverity: ErrorSeverity,
			expectedResults: LinterResultSet{
				{
					Severity: ErrorSeverity,
					Message:  `specification "spec2" at "" should have a version`,
				},
			},
		},
		{
			name: "multiple specifications are missing versions",
			when: specter.SpecificationGroup{
				&mockSpecification{name: "spec1", version: ""},
				&mockSpecification{name: "spec2", version: ""},
			},
			givenSeverity: ErrorSeverity,
			expectedResults: LinterResultSet{
				{
					Severity: ErrorSeverity,
					Message:  `specification "spec1" at "" should have a version`,
				},
				{
					Severity: ErrorSeverity,
					Message:  `specification "spec2" at "" should have a version`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := HasVersionMustHaveAVersionLinter(tt.givenSeverity)
			results := linter.Lint(tt.when)
			require.Equal(t, tt.expectedResults, results)
		})
	}
}