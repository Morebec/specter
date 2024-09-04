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

// Test cases for NewUnitGroup
func TestNewUnitGroup(t *testing.T) {
	tests := []struct {
		name  string
		given []Unit
		when  func() UnitGroup
		then  func(UnitGroup) bool
	}{
		{
			name:  "GIVEN no units WHEN calling NewUnitGroup THEN return an empty group",
			given: []Unit{},
			when: func() UnitGroup {
				return NewUnitGroup()
			},
			then: func(result UnitGroup) bool {
				return len(result) == 0
			},
		},
		{
			name: "GIVEN multiple units WHEN calling NewUnitGroup THEN return a group with those units",
			given: []Unit{
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			},
			when: func() UnitGroup {
				return NewUnitGroup(
					&UnitStub{name: "unit1", typeName: "type", source: Source{}},
					&UnitStub{name: "unit2", typeName: "type", source: Source{}},
				)
			},
			then: func(result UnitGroup) bool {
				return len(result) == 2 &&
					result[0].Name() == "unit1" &&
					result[1].Name() == "unit2"
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
func TestUnitGroup_Merge(t *testing.T) {
	tests := []struct {
		name  string
		given UnitGroup
		when  UnitGroup
		then  UnitGroup
	}{
		{
			name: "GIVEN two disjoint groups THEN return a group with all units",
			given: NewUnitGroup(
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
			),
			when: NewUnitGroup(
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			),
			then: NewUnitGroup(
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			),
		},
		{
			name: "GIVEN two groups with overlapping units THEN return a group without duplicates",
			given: NewUnitGroup(
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
			),
			when: NewUnitGroup(
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			),
			then: NewUnitGroup(
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
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

func TestUnitGroup_Select(t *testing.T) {
	tests := []struct {
		name  string
		given UnitGroup
		when  func(u Unit) bool
		then  UnitGroup
	}{
		{
			name: "GIVEN no units matches, THEN return an empty group",
			given: UnitGroup{
				&UnitStub{name: "unit2name", typeName: "type", source: Source{}},
			},
			when: func(u Unit) bool {
				return false
			},
			then: UnitGroup{},
		},
		{
			name: "GIVEN units matches, THEN return a group with only matching units",
			given: UnitGroup{
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			},
			when: func(u Unit) bool {
				return u.Name() == "unit2"
			},
			then: UnitGroup{
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
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

func TestUnitGroup_SelectType(t *testing.T) {
	tests := []struct {
		name  string
		given UnitGroup
		when  UnitType
		then  UnitGroup
	}{
		{
			name: "GIVEN no units matches, THEN return an empty group",
			given: UnitGroup{
				&UnitStub{name: "unit2name", typeName: "type", source: Source{}},
			},
			when: "not_found",
			then: UnitGroup{},
		},
		{
			name: "GIVEN a unit matches, THEN return a group with matching unit",
			given: UnitGroup{
				&UnitStub{name: "unit1", typeName: "type1", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type2", source: Source{}},
			},
			when: "type1",
			then: UnitGroup{
				&UnitStub{name: "unit1", typeName: "type1", source: Source{}},
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

func TestUnitGroup_SelectName(t *testing.T) {
	tests := []struct {
		name  string
		given UnitGroup
		when  UnitName
		then  Unit
	}{
		{
			name: "GIVEN a group with multiple units WHEN selecting an existing name THEN return the corresponding unit",
			given: NewUnitGroup(
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			),
			when: "unit2",
			then: &UnitStub{name: "unit2", typeName: "type", source: Source{}},
		},
		{
			name: "GIVEN a group with multiple units WHEN selecting a non-existent name THEN return nil",
			given: NewUnitGroup(
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			),
			when: "spec3",
			then: nil,
		},
		{
			name:  "GIVEN an empty group WHEN selecting a name THEN return nil",
			given: NewUnitGroup(),
			when:  "unit1",
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

func TestUnitGroup_SelectNames(t *testing.T) {
	tests := []struct {
		name  string
		given UnitGroup
		when  []UnitName
		then  UnitGroup
	}{
		{
			name: "GIVEN no units matches, THEN return a group with no values",
			given: UnitGroup{
				&UnitStub{name: "name", typeName: "type", source: Source{}},
			},
			when: []UnitName{"not_found"},
			then: UnitGroup{},
		},
		{
			name: "GIVEN a unit matches, THEN return a group with matching unit",
			given: UnitGroup{
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			},
			when: []UnitName{"unit1"},
			then: UnitGroup{
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
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

func TestUnitGroup_Exclude(t *testing.T) {
	tests := []struct {
		name  string
		given UnitGroup
		when  func(u Unit) bool
		then  UnitGroup
	}{
		{
			name: "GIVEN no units matches, THEN return a group with the same values",
			given: UnitGroup{
				&UnitStub{name: "name", typeName: "type", source: Source{}},
			},
			when: func(u Unit) bool {
				return false
			},
			then: UnitGroup{
				&UnitStub{name: "name", typeName: "type", source: Source{}},
			},
		},
		{
			name: "GIVEN units matches, THEN return a group without matching units",
			given: UnitGroup{
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			},
			when: func(u Unit) bool {
				return true
			},
			then: UnitGroup{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.Exclude(tt.when)
			require.Equal(t, tt.then, got)
		})
	}
}

func TestUnitGroup_ExcludeType(t *testing.T) {
	tests := []struct {
		name  string
		given UnitGroup
		when  UnitType
		then  UnitGroup
	}{
		{
			name: "GIVEN no units matches, THEN return a group with the same values",
			given: UnitGroup{
				&UnitStub{name: "unit2name", typeName: "type", source: Source{}},
			},
			when: "not_found",
			then: UnitGroup{
				&UnitStub{name: "unit2name", typeName: "type", source: Source{}},
			},
		},
		{
			name: "GIVEN a unit matches, THEN return a group without matching unit",
			given: UnitGroup{
				&UnitStub{name: "unit1", typeName: "type1", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type2", source: Source{}},
			},
			when: "type1",
			then: UnitGroup{
				&UnitStub{name: "unit2", typeName: "type2", source: Source{}},
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

func TestUnitGroup_ExcludeNames(t *testing.T) {
	tests := []struct {
		name  string
		given UnitGroup
		when  []UnitName
		then  UnitGroup
	}{
		{
			name: "GIVEN no units matches, THEN return a group with the same values",
			given: UnitGroup{
				&UnitStub{name: "unit2name", typeName: "type", source: Source{}},
			},
			when: []UnitName{"not_found"},
			then: UnitGroup{
				&UnitStub{name: "unit2name", typeName: "type", source: Source{}},
			},
		},
		{
			name: "GIVEN a unit matches, THEN return a group without matching unit",
			given: UnitGroup{
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			},
			when: []UnitName{"unit1"},
			then: UnitGroup{
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
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

func TestMapUnitGroup(t *testing.T) {
	tests := []struct {
		name  string
		given UnitGroup
		when  func(Unit) string
		then  []string
	}{
		{
			name: "GIVEN a group with multiple units WHEN mapped to their names THEN return a slice of unit names",
			given: NewUnitGroup(
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			),
			when: func(u Unit) string {
				return string(u.Name())
			},
			then: []string{"unit1", "unit2"},
		},
		{
			name:  "GIVEN an empty group WHEN mapped THEN return a nil slice",
			given: NewUnitGroup(),
			when: func(u Unit) string {
				return string(u.Name())
			},
			then: nil,
		},
		{
			name: "GIVEN a group with multiple units WHEN mapped to a constant value THEN return a slice of that value",
			given: NewUnitGroup(
				&UnitStub{name: "unit1", typeName: "type", source: Source{}},
				&UnitStub{name: "unit2", typeName: "type", source: Source{}},
			),
			when: func(u Unit) string {
				return "constant"
			},
			then: []string{"constant", "constant"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapUnitGroup(tt.given, tt.when)
			assert.Equal(t, tt.then, got)
		})
	}
}
