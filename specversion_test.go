package specter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Mock implementations for the Specification and HasVersion interfaces

var _ Specification = (*mockSpecification)(nil)

type mockSpecification struct {
	name        SpecificationName
	description string
	source      Source
	version     SpecificationVersion
	typeName    SpecificationType
}

func (m *mockSpecification) Name() SpecificationName {
	return m.name
}

func (m *mockSpecification) Type() SpecificationType {
	return m.typeName
}

func (m *mockSpecification) Description() string {
	return m.description
}

func (m *mockSpecification) SetSource(s Source) {
	m.source = s
}

func (m *mockSpecification) Source() Source {
	return m.source
}

func (m *mockSpecification) Version() SpecificationVersion {
	return m.version
}

func TestHasVersionMustHaveAVersionLinter(t *testing.T) {
	tests := []struct {
		name            string
		given           SpecificationGroup
		expectedResults LinterResultSet
		severity        LinterResultSeverity
	}{
		{
			name: "GIVEN all specifications have a version THEN return no warnings or errors",
			given: SpecificationGroup{
				&mockSpecification{name: "spec1", version: "v1"},
				&mockSpecification{name: "spec2", version: "v2"},
			},
			severity:        WarningSeverity,
			expectedResults: LinterResultSet(nil),
		},
		{
			name: "GIVEN one specification is missing a version and severity is Warning THEN return a warning",
			given: SpecificationGroup{
				&mockSpecification{name: "spec1", version: "v1"},
				&mockSpecification{name: "spec2", version: ""},
			},
			severity: WarningSeverity,
			expectedResults: LinterResultSet{
				{
					Severity: WarningSeverity,
					Message:  `specification "spec2" at "" should have a version`,
				},
			},
		},
		{
			name: "GIVEN one specification is missing a version and severity is error THEN return an error",
			given: SpecificationGroup{
				&mockSpecification{name: "spec1", version: "v1"},
				&mockSpecification{name: "spec2", version: ""},
			},
			severity: ErrorSeverity,
			expectedResults: LinterResultSet{
				{
					Severity: ErrorSeverity,
					Message:  `specification "spec2" at "" should have a version`,
				},
			},
		},
		{
			name: "multiple specifications are missing versions",
			given: SpecificationGroup{
				&mockSpecification{name: "spec1", version: ""},
				&mockSpecification{name: "spec2", version: ""},
			},
			severity: ErrorSeverity,
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
			linter := HasVersionMustHaveAVersionLinter(tt.severity)
			results := linter.Lint(tt.given)
			assert.Equal(t, tt.expectedResults, results)
		})
	}
}
