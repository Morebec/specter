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
	"github.com/morebec/specter/pkg/specter"
	"github.com/morebec/specter/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// Test cases for NewUnitGroup
func TestNewUnitGroup(t *testing.T) {
	tests := []struct {
		name  string
		given []specter.Unit
		when  func() specter.UnitGroup
		then  func(specter.UnitGroup) bool
	}{
		{
			name:  "GIVEN no units WHEN calling NewUnitGroup THEN return an empty group",
			given: []specter.Unit{},
			when: func() specter.UnitGroup {
				return specter.NewUnitGroup()
			},
			then: func(result specter.UnitGroup) bool {
				return len(result) == 0
			},
		},
		{
			name: "GIVEN multiple units WHEN calling NewUnitGroup THEN return a group with those units",
			given: []specter.Unit{
				testutils.NewUnitStub("unit1", "type", specter.Source{}),
				testutils.NewUnitStub("unit2", "type", specter.Source{}),
			},
			when: func() specter.UnitGroup {
				return specter.NewUnitGroup(
					testutils.NewUnitStub("unit1", "type", specter.Source{}),
					testutils.NewUnitStub("unit2", "type", specter.Source{}),
				)
			},
			then: func(result specter.UnitGroup) bool {
				return len(result) == 2 &&
					result[0].ID() == "unit1" &&
					result[1].ID() == "unit2"
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
		given specter.UnitGroup
		when  specter.UnitGroup
		then  specter.UnitGroup
	}{
		{
			name: "GIVEN two disjoint groups THEN return a group with all units",
			given: specter.NewUnitGroup(
				testutils.NewUnitStub("unit1", "type", specter.Source{}),
			),
			when: specter.NewUnitGroup(
				testutils.NewUnitStub("unit2", "type", specter.Source{}),
			),
			then: specter.NewUnitGroup(
				testutils.NewUnitStub("unit1", "type", specter.Source{}),
				testutils.NewUnitStub("unit2", "type", specter.Source{}),
			),
		},
		{
			name: "GIVEN two groups with overlapping units THEN return a group without duplicates",
			given: specter.NewUnitGroup(
				testutils.NewUnitStub("unit1", "type", specter.Source{}),
			),
			when: specter.NewUnitGroup(
				testutils.NewUnitStub("unit1", "type", specter.Source{}),
				testutils.NewUnitStub("unit2", "type", specter.Source{}),
			),
			then: specter.NewUnitGroup(
				testutils.NewUnitStub("unit1", "type", specter.Source{}),
				testutils.NewUnitStub("unit2", "type", specter.Source{}),
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
		given specter.UnitGroup
		when  func(u specter.Unit) bool
		then  specter.UnitGroup
	}{
		{
			name: "GIVEN no units matches, THEN return an empty group",
			given: specter.UnitGroup{
				testutils.NewUnitStub("unit2name", "type", specter.Source{}),
			},
			when: func(u specter.Unit) bool {
				return false
			},
			then: specter.UnitGroup{},
		},
		{
			name: "GIVEN units matches, THEN return a group with only matching units",
			given: specter.UnitGroup{
				testutils.NewUnitStub("unit1", "type", specter.Source{}),
				testutils.NewUnitStub("unit2", "type", specter.Source{}),
			},
			when: func(u specter.Unit) bool {
				return u.ID() == "unit2"
			},
			then: specter.UnitGroup{
				testutils.NewUnitStub("unit2", "type", specter.Source{}),
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

func TestUnitGroup_Exclude(t *testing.T) {
	tests := []struct {
		name  string
		given specter.UnitGroup
		when  func(u specter.Unit) bool
		then  specter.UnitGroup
	}{
		{
			name: "GIVEN no units matches, THEN return a group with the same values",
			given: specter.UnitGroup{
				testutils.NewUnitStub("name", "type", specter.Source{}),
			},
			when: func(u specter.Unit) bool {
				return false
			},
			then: specter.UnitGroup{
				testutils.NewUnitStub("name", "type", specter.Source{}),
			},
		},
		{
			name: "GIVEN units matches, THEN return a group without matching units",
			given: specter.UnitGroup{
				testutils.NewUnitStub("unit1", "type", specter.Source{}),
				testutils.NewUnitStub("unit2", "type", specter.Source{}),
			},
			when: func(u specter.Unit) bool {
				return true
			},
			then: specter.UnitGroup{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.Exclude(tt.when)
			require.Equal(t, tt.then, got)
		})
	}
}

func TestMapUnitGroup(t *testing.T) {
	tests := []struct {
		name  string
		given specter.UnitGroup
		when  func(specter.Unit) string
		then  []string
	}{
		{
			name: "GIVEN a group with multiple units WHEN mapped to their names THEN return a slice of unit names",
			given: specter.NewUnitGroup(
				testutils.NewUnitStub("unit1", "type", specter.Source{}),
				testutils.NewUnitStub("unit2", "type", specter.Source{}),
			),
			when: func(u specter.Unit) string {
				return string(u.ID())
			},
			then: []string{"unit1", "unit2"},
		},
		{
			name:  "GIVEN an empty group WHEN mapped THEN return a nil slice",
			given: specter.NewUnitGroup(),
			when: func(u specter.Unit) string {
				return string(u.ID())
			},
			then: nil,
		},
		{
			name: "GIVEN a group with multiple units WHEN mapped to a constant value THEN return a slice of that value",
			given: specter.NewUnitGroup(
				testutils.NewUnitStub("unit1", "type", specter.Source{}),
				testutils.NewUnitStub("unit2", "type", specter.Source{}),
			),
			when: func(u specter.Unit) string {
				return "constant"
			},
			then: []string{"constant", "constant"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := specter.MapUnitGroup(tt.given, tt.when)
			assert.Equal(t, tt.then, got)
		})
	}
}

func TestUnitGroup_Find(t *testing.T) {
	type then struct {
		unit  specter.Unit
		found bool
	}
	tests := []struct {
		name  string
		given specter.UnitGroup
		when  specter.UnitMatcher
		then  then
	}{
		{
			name: "WHEN a unit matches THEN should return unit and true",
			given: specter.UnitGroup{
				testutils.NewUnitStub("", "", specter.Source{}),
			},
			when: func(u specter.Unit) bool {
				return true
			},
			then: then{
				unit:  testutils.NewUnitStub("", "", specter.Source{}),
				found: true,
			},
		},
		{
			name: "WHEN no unit matches THEN should return nil and false",
			given: specter.UnitGroup{
				testutils.NewUnitStub("", "", specter.Source{}),
			},
			when: func(u specter.Unit) bool {
				return false
			},
			then: then{
				unit:  nil,
				found: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := tt.given.Find(tt.when)
			require.Equal(t, tt.then.found, found)
			require.Equal(t, tt.then.unit, got)
		})
	}
}

func TestUnitOf(t *testing.T) {
	t.Run("Attributes should be set to passed values", func(t *testing.T) {
		value := "hello-world"
		var unitID specter.UnitID = "unitID"
		var unitKind specter.UnitKind = "kind"
		unitSource := specter.Source{
			Location: "/path/to/file",
			Data:     []byte(`some data`),
			Format:   "txt",
		}
		u := specter.UnitOf(value, unitID, unitKind, unitSource)
		require.NotNil(t, u)
		require.Equal(t, unitID, u.ID())
		require.Equal(t, unitKind, u.Kind())
		require.Equal(t, unitSource, u.Source())
		require.Equal(t, value, u.Unwrap())
	})
}

func TestUnitWithKindMatcher(t *testing.T) {
	type when struct {
		kind specter.UnitKind
		unit specter.Unit
	}
	tests := []struct {
		name string
		when when
		then bool
	}{
		{
			name: "WHEN unit with kind THEN return true",
			when: when{
				kind: "kind",
				unit: testutils.NewUnitStub("unit1", "kind", specter.Source{}),
			},
			then: true,
		},
		{
			name: "WHEN unit not with kind THEN return false",
			when: when{
				kind: "kind",
				unit: testutils.NewUnitStub("unit1", "not_kind", specter.Source{}),
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := specter.UnitWithKindsMatcher(tt.when.kind)(tt.when.unit)
			assert.Equal(t, tt.then, got)
		})
	}
}

func TestUnitWithIDsMatcher(t *testing.T) {
	type when struct {
		id   specter.UnitID
		unit specter.Unit
	}
	tests := []struct {
		name string
		when when
		then bool
	}{
		{
			name: "WHEN unit with id THEN return true",
			when: when{
				id:   "unit1",
				unit: testutils.NewUnitStub("unit1", "kind", specter.Source{}),
			},
			then: true,
		},
		{
			name: "WHEN unit not with id THEN return false",
			when: when{
				id:   "id",
				unit: testutils.NewUnitStub("not_id", "kind", specter.Source{}),
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := specter.UnitWithIDsMatcher(tt.when.id)(tt.when.unit)
			assert.Equal(t, tt.then, got)
		})
	}
}

func TestUnitLoaderAdapter(t *testing.T) {
	t.Run("functions should be called", func(t *testing.T) {
		expectedSource := specter.Source{
			Location: "/path/to/file",
		}
		expectedUnits := []specter.Unit{
			specter.UnitOf(0, "", "", expectedSource),
		}

		f := specter.UnitLoaderAdapter{
			LoadFunc: func(s specter.Source) ([]specter.Unit, error) {
				assert.Equal(t, expectedSource, s)
				return expectedUnits, nil
			},
			SupportsSourceFunc: func(s specter.Source) bool {
				assert.Equal(t, expectedSource, s)
				return true
			},
		}
		assert.True(t, f.SupportsSource(expectedSource))

		units, err := f.Load(expectedSource)
		require.NoError(t, err)
		assert.Equal(t, expectedUnits, units)
	})
	t.Run("functions should not fail if no parameters provided", func(t *testing.T) {
		f := specter.UnitLoaderAdapter{}
		assert.False(t, f.SupportsSource(specter.Source{}))

		units, err := f.Load(specter.Source{})
		require.NoError(t, err)
		assert.Nil(t, units)
	})
}

func TestUnwrapUnit(t *testing.T) {
	type then struct {
		value any
		ok    bool
	}
	type testCase struct {
		name string
		when specter.Unit
		then then
	}
	tests := []testCase{
		{
			name: "Unwrap of non wrapped unit should return zero value and false",
			when: testutils.NewUnitStub("id", "kind", specter.Source{}),
			then: then{
				value: "",
				ok:    false,
			},
		},
		{
			name: "Unwrap of a wrapped unit should return the value and true",
			when: specter.UnitOf("hello", "id", "kind", specter.Source{}),
			then: then{
				value: "hello",
				ok:    true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := specter.UnwrapUnit[string](tt.when)
			assert.Equal(t, tt.then.value, gotValue)
			assert.Equal(t, tt.then.ok, gotOk)
		})
	}
}
