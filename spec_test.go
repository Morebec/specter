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

package specter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"testing"
)

func TestGenericSpecification_Description(t *testing.T) {
	tests := []struct {
		name  string
		given *GenericSpecification
		then  string
	}{
		{
			name: "GIVEN a specification with a description attribute THEN return the description",
			given: &GenericSpecification{
				Attributes: []GenericSpecAttribute{
					{
						Name:  "description",
						Value: GenericValue{cty.StringVal("This is a test specification")},
					},
				},
			},
			then: "This is a test specification",
		},
		{
			name: "GIVEN a specification without a description attribute THEN return an empty string",
			given: &GenericSpecification{
				Attributes: []GenericSpecAttribute{},
			},
			then: "",
		},
		{
			name: "GIVEN a specification with a non-string description THEN return an empty string",
			given: &GenericSpecification{
				Attributes: []GenericSpecAttribute{
					{
						Name:  "description",
						Value: GenericValue{cty.NumberIntVal(42)}, // Not a string value
					},
				},
			},
			then: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.Description()
			assert.Equal(t, tt.then, got)
		})
	}
}

func TestGenericSpecification_Attribute(t *testing.T) {
	tests := []struct {
		name  string
		given *GenericSpecification
		when  string
		then  *GenericSpecAttribute
	}{
		{
			name: "GIVEN a specification with a specific attribute WHEN Attribute is called THEN return the attribute",
			given: &GenericSpecification{
				Attributes: []GenericSpecAttribute{
					{
						Name:  "attr1",
						Value: GenericValue{cty.StringVal("value1")},
					},
				},
			},
			when: "attr1",
			then: &GenericSpecAttribute{
				Name:  "attr1",
				Value: GenericValue{cty.StringVal("value1")},
			},
		},
		{
			name: "GIVEN a specification without the specified attribute WHEN Attribute is called THEN return nil",
			given: &GenericSpecification{
				Attributes: []GenericSpecAttribute{},
			},
			when: "nonexistent",
			then: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.Attribute(tt.when)
			assert.Equal(t, tt.then, got)
		})
	}
}

func TestGenericSpecification_HasAttribute(t *testing.T) {
	tests := []struct {
		name  string
		given *GenericSpecification
		when  string
		then  bool
	}{
		{
			name: "GIVEN a specification with a specific attribute WHEN HasAttribute is called THEN return true",
			given: &GenericSpecification{
				Attributes: []GenericSpecAttribute{
					{
						Name:  "attr1",
						Value: GenericValue{cty.StringVal("value1")},
					},
				},
			},
			when: "attr1",
			then: true,
		},
		{
			name: "GIVEN a specification without the specified attribute WHEN HasAttribute is called THEN return false",
			given: &GenericSpecification{
				Attributes: []GenericSpecAttribute{},
			},
			when: "nonexistent",
			then: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.HasAttribute(tt.when)
			assert.Equal(t, tt.then, got)
		})
	}
}

func TestGenericSpecification_SetSource(t *testing.T) {
	tests := []struct {
		name  string
		given *GenericSpecification
		when  Source
		then  Source
	}{
		{
			name: "GIVEN a specification WHEN SetSource is called THEN updates the source",
			given: &GenericSpecification{
				source: Source{Location: "initial/path"},
			},
			when: Source{Location: "new/path"},
			then: Source{Location: "new/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.given.SetSource(tt.when)
			assert.Equal(t, tt.then, tt.given.Source())
		})
	}
}

// Test cases for NewSpecGroup
func TestNewSpecGroup(t *testing.T) {
	tests := []struct {
		name  string
		given []Specification
		when  func() SpecificationGroup
		then  func(SpecificationGroup) bool
	}{
		{
			name:  "GIVEN no specifications WHEN calling NewSpecGroup THEN return an empty group",
			given: []Specification{},
			when: func() SpecificationGroup {
				return NewSpecGroup()
			},
			then: func(result SpecificationGroup) bool {
				return len(result) == 0
			},
		},
		{
			name: "GIVEN multiple specifications WHEN calling NewSpecGroup THEN return a group with those specifications",
			given: []Specification{
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
			},
			when: func() SpecificationGroup {
				return NewSpecGroup(
					NewGenericSpecification("spec1", "type", Source{}),
					NewGenericSpecification("spec2", "type", Source{}),
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

// Test cases for merge
func TestSpecificationGroup_Merge(t *testing.T) {
	tests := []struct {
		name  string
		given SpecificationGroup
		when  SpecificationGroup
		then  SpecificationGroup
	}{
		{
			name: "GIVEN two disjoint groups THEN return a group with all specifications",
			given: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}),
			),
			when: NewSpecGroup(
				NewGenericSpecification("spec2", "type", Source{}),
			),
			then: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
			),
		},
		{
			name: "GIVEN two groups with overlapping specifications THEN return a group without duplicates",
			given: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}),
			),
			when: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
			),
			then: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
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
				NewGenericSpecification("name", "type", Source{}),
			},
			when: func(s Specification) bool {
				return false
			},
			then: SpecificationGroup{},
		},
		{
			name: "GIVEN specifications matches, THEN return a group with only matching specifications",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
			},
			when: func(s Specification) bool {
				return s.Name() == "spec2"
			},
			then: SpecificationGroup{
				NewGenericSpecification("spec2", "type", Source{}),
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
				NewGenericSpecification("name", "type", Source{}),
			},
			when: "not_found",
			then: SpecificationGroup{},
		},
		{
			name: "GIVEN a specification matches, THEN return a group with matching specification",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type1", Source{}),
				NewGenericSpecification("spec2", "type2", Source{}),
			},
			when: "type1",
			then: SpecificationGroup{
				NewGenericSpecification("spec1", "type1", Source{}),
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

func TestSpecificationGroup_SelectName(t *testing.T) {
	tests := []struct {
		name  string
		given SpecificationGroup
		when  SpecificationName
		then  Specification
	}{
		{
			name: "GIVEN a group with multiple specifications WHEN selecting an existing name THEN return the corresponding specification",
			given: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
			),
			when: "spec2",
			then: NewGenericSpecification("spec2", "type", Source{}),
		},
		{
			name: "GIVEN a group with multiple specifications WHEN selecting a non-existent name THEN return nil",
			given: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
			),
			when: "spec3",
			then: nil,
		},
		{
			name:  "GIVEN an empty group WHEN selecting a name THEN return nil",
			given: NewSpecGroup(),
			when:  "spec1",
			then:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.SelectName(tt.when)
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
				NewGenericSpecification("name", "type", Source{}),
			},
			when: []SpecificationName{"not_found"},
			then: SpecificationGroup{},
		},
		{
			name: "GIVEN a specification matches, THEN return a group with matching specification",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
			},
			when: []SpecificationName{"spec1"},
			then: SpecificationGroup{
				NewGenericSpecification("spec1", "type", Source{}),
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
				NewGenericSpecification("name", "type", Source{}),
			},
			when: func(s Specification) bool {
				return false
			},
			then: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}),
			},
		},
		{
			name: "GIVEN specifications matches, THEN return a group without matching specifications",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
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
				NewGenericSpecification("name", "type", Source{}),
			},
			when: "not_found",
			then: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}),
			},
		},
		{
			name: "GIVEN a specification matches, THEN return a group without matching specification",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type1", Source{}),
				NewGenericSpecification("spec2", "type2", Source{}),
			},
			when: "type1",
			then: SpecificationGroup{
				NewGenericSpecification("spec2", "type2", Source{}),
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
				NewGenericSpecification("name", "type", Source{}),
			},
			when: []SpecificationName{"not_found"},
			then: SpecificationGroup{
				NewGenericSpecification("name", "type", Source{}),
			},
		},
		{
			name: "GIVEN a specification matches, THEN return a group without matching specification",
			given: SpecificationGroup{
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
			},
			when: []SpecificationName{"spec1"},
			then: SpecificationGroup{
				NewGenericSpecification("spec2", "type", Source{}),
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
			name: "GIVEN a group with multiple specifications WHEN mapped to their names THEN return a slice of specification names",
			given: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
			),
			when: func(s Specification) string {
				return string(s.Name())
			},
			then: []string{"spec1", "spec2"},
		},
		{
			name:  "GIVEN an empty group WHEN mapped THEN return a nil slice",
			given: NewSpecGroup(),
			when: func(s Specification) string {
				return string(s.Name())
			},
			then: nil,
		},
		{
			name: "GIVEN a group with multiple specifications WHEN mapped to a constant value THEN return a slice of that value",
			given: NewSpecGroup(
				NewGenericSpecification("spec1", "type", Source{}),
				NewGenericSpecification("spec2", "type", Source{}),
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

func TestObjectValue_String(t *testing.T) {
	o := ObjectValue{Type: "hello", Attributes: []GenericSpecAttribute{
		{Name: "hello", Value: GenericValue{cty.StringVal("world")}},
	}}
	assert.Equal(t, "ObjectValue{Type: hello, Attributes: [{hello world}]}", o.String())
}
