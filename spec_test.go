package specter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// Test cases for NewSpecGroup
func TestNewSpecGroup(t *testing.T) {
	tests := []struct {
		name  string
		given []Specification
		when  func() SpecificationGroup
		then  func(SpecificationGroup) bool
	}{
		{
			name:  "GIVEN no specifications WHEN calling NewSpecGroup THEN returns an empty group",
			given: []Specification{},
			when: func() SpecificationGroup {
				return NewSpecGroup()
			},
			then: func(result SpecificationGroup) bool {
				return len(result) == 0
			},
		},
		{
			name: "GIVEN multiple specifications WHEN calling NewSpecGroup THEN returns a group with those specifications",
			given: []Specification{
				NewGenericSpecification("spec1", "type", Source{}, nil),
				NewGenericSpecification("spec2", "type", Source{}, nil),
			},
			when: func() SpecificationGroup {
				return NewSpecGroup(
					NewGenericSpecification("spec1", "type", Source{}, nil),
					NewGenericSpecification("spec2", "type", Source{}, nil),
				)
			},
			then: func(result SpecificationGroup) bool {
				return len(result) == 2 &&
					result[0].Name() == "spec1" &&
					result[1].Name() == "spec2"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.when()
			if !tt.then(result) {
				t.Errorf("Test %s failed", tt.name)
			}
		})
	}
}

// Test cases for Merge
func TestSpecificationGroup_Merge(t *testing.T) {
	tests := []struct {
		name  string
		given SpecificationGroup
		when  SpecificationGroup
		then  SpecificationGroup
	}{
		{
			name: "GIVEN two disjoint groups THEN returns a group with all specifications",
			given: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}, nil),
			),
			when: NewSpecGroup(
				NewGenericSpecification("spec2", "type", Source{}, nil),
			),
			then: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}, nil),
				NewGenericSpecification("spec2", "type", Source{}, nil),
			),
		},
		{
			name: "GIVEN two groups with overlapping specifications THEN returns a group without duplicates",
			given: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}, nil),
			),
			when: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}, nil),
				NewGenericSpecification("spec2", "type", Source{}, nil),
			),
			then: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}, nil),
				NewGenericSpecification("spec2", "type", Source{}, nil),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.Merge(tt.when)
			require.Equal(t, tt.then, got)
		})
	}
}

func TestSpecificationGroup_Select(t *testing.T) {
	tests := []struct {
		name  string
		given SpecificationGroup
		when  func(s Specification) bool
		then  SpecificationGroup
	}{
		{
			name: "GIVEN no specifications matches, THEN return an empty group",
			given: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}, nil),
			},
			when: func(s Specification) bool {
				return false
			},
			then: SpecificationGroup{},
		},
		{
			name: "GIVEN specifications matches, THEN return a group with only matching specifications",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type", Source{}, nil),
				NewGenericSpecification("spec2", "type", Source{}, nil),
			},
			when: func(s Specification) bool {
				return s.Name() == "spec2"
			},
			then: SpecificationGroup{
				NewGenericSpecification("spec2", "type", Source{}, nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.Select(tt.when)
			require.Equal(t, tt.then, got)
		})
	}
}

func TestSpecificationGroup_SelectType(t *testing.T) {
	tests := []struct {
		name  string
		given SpecificationGroup
		when  SpecificationType
		then  SpecificationGroup
	}{
		{
			name: "GIVEN no specifications matches, THEN return an empty group",
			given: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}, nil),
			},
			when: "not_found",
			then: SpecificationGroup{},
		},
		{
			name: "GIVEN a specification matches, THEN return a group with matching specification",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type1", Source{}, nil),
				NewGenericSpecification("spec2", "type2", Source{}, nil),
			},
			when: "type1",
			then: SpecificationGroup{
				NewGenericSpecification("spec1", "type1", Source{}, nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.SelectType(tt.when)
			require.Equal(t, tt.then, got)
		})
	}
}

func TestSpecificationGroup_SelectNames(t *testing.T) {
	tests := []struct {
		name  string
		given SpecificationGroup
		when  []SpecificationName
		then  SpecificationGroup
	}{
		{
			name: "GIVEN no specifications matches, THEN return a group with no values",
			given: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}, nil),
			},
			when: []SpecificationName{"not_found"},
			then: SpecificationGroup{},
		},
		{
			name: "GIVEN a specification matches, THEN return a group with matching specification",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type", Source{}, nil),
				NewGenericSpecification("spec2", "type", Source{}, nil),
			},
			when: []SpecificationName{"spec1"},
			then: SpecificationGroup{
				NewGenericSpecification("spec1", "type", Source{}, nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.SelectNames(tt.when...)
			require.Equal(t, tt.then, got)
		})
	}
}

func TestSpecificationGroup_Exclude(t *testing.T) {
	tests := []struct {
		name  string
		given SpecificationGroup
		when  func(s Specification) bool
		then  SpecificationGroup
	}{
		{
			name: "GIVEN no specifications matches, THEN return a group with the same values",
			given: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}, nil),
			},
			when: func(s Specification) bool {
				return false
			},
			then: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}, nil),
			},
		},
		{
			name: "GIVEN specifications matches, THEN return a group without matching specifications",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type", Source{}, nil),
				NewGenericSpecification("spec2", "type", Source{}, nil),
			},
			when: func(s Specification) bool {
				return true
			},
			then: SpecificationGroup{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.Exclude(tt.when)
			require.Equal(t, tt.then, got)
		})
	}
}

func TestSpecificationGroup_ExcludeType(t *testing.T) {
	tests := []struct {
		name  string
		given SpecificationGroup
		when  SpecificationType
		then  SpecificationGroup
	}{
		{
			name: "GIVEN no specifications matches, THEN return a group with the same values",
			given: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}, nil),
			},
			when: "not_found",
			then: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}, nil),
			},
		},
		{
			name: "GIVEN a specification matches, THEN return a group without matching specification",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type1", Source{}, nil),
				NewGenericSpecification("spec2", "type2", Source{}, nil),
			},
			when: "type1",
			then: SpecificationGroup{
				NewGenericSpecification("spec2", "type2", Source{}, nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.ExcludeType(tt.when)
			require.Equal(t, tt.then, got)
		})
	}
}

func TestSpecificationGroup_ExcludeNames(t *testing.T) {
	tests := []struct {
		name  string
		given SpecificationGroup
		when  []SpecificationName
		then  SpecificationGroup
	}{
		{
			name: "GIVEN no specifications matches, THEN return a group with the same values",
			given: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}, nil),
			},
			when: []SpecificationName{"not_found"},
			then: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}, nil),
			},
		},
		{
			name: "GIVEN a specification matches, THEN return a group without matching specification",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type", Source{}, nil),
				NewGenericSpecification("spec2", "type", Source{}, nil),
			},
			when: []SpecificationName{"spec1"},
			then: SpecificationGroup{
				NewGenericSpecification("spec2", "type", Source{}, nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.ExcludeNames(tt.when...)
			require.Equal(t, tt.then, got)
		})
	}
}

func TestMapSpecGroup(t *testing.T) {
	tests := []struct {
		name  string
		given SpecificationGroup
		when  func(Specification) string
		then  []string
	}{
		{
			name: "GIVEN a group with multiple specifications WHEN mapped to their names THEN returns a slice of specification names",
			given: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}, nil),
				NewGenericSpecification("spec2", "type", Source{}, nil),
			),
			when: func(s Specification) string {
				return string(s.Name())
			},
			then: []string{"spec1", "spec2"},
		},
		{
			name:  "GIVEN an empty group WHEN mapped THEN returns a nil slice",
			given: NewSpecGroup(),
			when: func(s Specification) string {
				return string(s.Name())
			},
			then: nil,
		},
		{
			name: "GIVEN a group with multiple specifications WHEN mapped to a constant value THEN returns a slice of that value",
			given: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}, nil),
				NewGenericSpecification("spec2", "type", Source{}, nil),
			),
			when: func(s Specification) string {
				return "constant"
			},
			then: []string{"constant", "constant"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapSpecGroup(tt.given, tt.when)
			assert.Equal(t, tt.then, got)
		})
	}
}
