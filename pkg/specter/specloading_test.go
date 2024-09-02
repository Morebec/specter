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

package specter_test

import (
	. "github.com/morebec/specter/pkg/specter"
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
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
			},
			when: func() SpecificationGroup {
				return NewSpecGroup(
					&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
					&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
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
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
			),
			when: NewSpecGroup(
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
			),
			then: NewSpecGroup(
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
			),
		},
		{
			name: "GIVEN two groups with overlapping specifications THEN return a group without duplicates",
			given: NewSpecGroup(
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
			),
			when: NewSpecGroup(
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
			),
			then: NewSpecGroup(
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
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
				&SpecificationStub{name: "spec2name", typeName: "type", source: Source{}},
			},
			when: func(s Specification) bool {
				return false
			},
			then: SpecificationGroup{},
		},
		{
			name: "GIVEN specifications matches, THEN return a group with only matching specifications",
			given: SpecificationGroup{
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
			},
			when: func(s Specification) bool {
				return s.Name() == "spec2"
			},
			then: SpecificationGroup{
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
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
				&SpecificationStub{name: "spec2name", typeName: "type", source: Source{}},
			},
			when: "not_found",
			then: SpecificationGroup{},
		},
		{
			name: "GIVEN a specification matches, THEN return a group with matching specification",
			given: SpecificationGroup{
				&SpecificationStub{name: "spec1", typeName: "type1", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type2", source: Source{}},
			},
			when: "type1",
			then: SpecificationGroup{
				&SpecificationStub{name: "spec1", typeName: "type1", source: Source{}},
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
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
			),
			when: "spec2",
			then: &SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
		},
		{
			name: "GIVEN a group with multiple specifications WHEN selecting a non-existent name THEN return nil",
			given: NewSpecGroup(
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
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
				&SpecificationStub{name: "name", typeName: "type", source: Source{}},
			},
			when: []SpecificationName{"not_found"},
			then: SpecificationGroup{},
		},
		{
			name: "GIVEN a specification matches, THEN return a group with matching specification",
			given: SpecificationGroup{
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
			},
			when: []SpecificationName{"spec1"},
			then: SpecificationGroup{
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
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
				&SpecificationStub{name: "name", typeName: "type", source: Source{}},
			},
			when: func(s Specification) bool {
				return false
			},
			then: SpecificationGroup{
				&SpecificationStub{name: "name", typeName: "type", source: Source{}},
			},
		},
		{
			name: "GIVEN specifications matches, THEN return a group without matching specifications",
			given: SpecificationGroup{
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
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
				&SpecificationStub{name: "spec2name", typeName: "type", source: Source{}},
			},
			when: "not_found",
			then: SpecificationGroup{
				&SpecificationStub{name: "spec2name", typeName: "type", source: Source{}},
			},
		},
		{
			name: "GIVEN a specification matches, THEN return a group without matching specification",
			given: SpecificationGroup{
				&SpecificationStub{name: "spec1", typeName: "type1", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type2", source: Source{}},
			},
			when: "type1",
			then: SpecificationGroup{
				&SpecificationStub{name: "spec2", typeName: "type2", source: Source{}},
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
				&SpecificationStub{name: "spec2name", typeName: "type", source: Source{}},
			},
			when: []SpecificationName{"not_found"},
			then: SpecificationGroup{
				&SpecificationStub{name: "spec2name", typeName: "type", source: Source{}},
			},
		},
		{
			name: "GIVEN a specification matches, THEN return a group without matching specification",
			given: SpecificationGroup{
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
			},
			when: []SpecificationName{"spec1"},
			then: SpecificationGroup{
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
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
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
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
				&SpecificationStub{name: "spec1", typeName: "type", source: Source{}},
				&SpecificationStub{name: "spec2", typeName: "type", source: Source{}},
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
