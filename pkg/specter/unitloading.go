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

// UnsupportedSourceErrorCode ErrorSeverity code returned by a UnitLoader when a given loader does not support a certain source.
const UnsupportedSourceErrorCode = "specter.spec_loading.unsupported_source"

// UnitLoader is a service responsible for loading Units from Sources.
type UnitLoader interface {
	// Load loads a slice of Unit from a Source, or returns an error if it encountered a failure.
	Load(s Source) ([]Unit, error)

	// SupportsSource indicates if this loader supports a certain source or not.
	SupportsSource(s Source) bool
}

type UnitType string

type UnitName string

// Unit is a general purpose data structure to represent a unit as loaded from a file regardless of the loader
// used.
// It is the responsibility of the application using specter to convert a unit to an appropriate data structure representing the intent of a
// given Unit.
type Unit interface {
	// Name returns the unique Name of this unit.
	Name() UnitName

	// Type returns the type of this unit.
	Type() UnitType

	// Description of this unit.
	Description() string

	// Source returns the source of this unit.
	Source() Source

	// SetSource sets the source of the unit.
	// This method should only be used by loaders.
	SetSource(s Source)
}

// UnitGroup Represents a list of Unit.
type UnitGroup []Unit

func NewUnitGroup(u ...Unit) UnitGroup {
	g := UnitGroup{}
	return append(g, u...)
}

// Merge Allows merging a group with another one.
func (g UnitGroup) Merge(group UnitGroup) UnitGroup {
	merged := g
	typeNameIndex := map[UnitName]any{}
	for _, u := range g {
		typeNameIndex[u.Name()] = nil
	}
	for _, u := range group {
		if _, found := typeNameIndex[u.Name()]; found {
			continue
		}
		typeNameIndex[u.Name()] = nil
		merged = append(merged, u)
	}
	return merged
}

// Select allows filtering the group for certain units.
func (g UnitGroup) Select(p func(u Unit) bool) UnitGroup {
	r := UnitGroup{}
	for _, u := range g {
		if p(u) {
			r = append(r, u)
		}
	}

	return r
}

func (g UnitGroup) SelectType(t UnitType) UnitGroup {
	return g.Select(func(u Unit) bool {
		return u.Type() == t
	})
}

func (g UnitGroup) SelectName(t UnitName) Unit {
	for _, u := range g {
		if u.Name() == t {
			return u
		}
	}

	return nil
}

func (g UnitGroup) SelectNames(names ...UnitName) UnitGroup {
	return g.Select(func(u Unit) bool {
		for _, name := range names {
			if u.Name() == name {
				return true
			}
		}
		return false
	})
}

func (g UnitGroup) Exclude(p func(u Unit) bool) UnitGroup {
	r := UnitGroup{}
	for _, u := range g {
		if !p(u) {
			r = append(r, u)
		}
	}

	return r
}

func (g UnitGroup) ExcludeType(t UnitType) UnitGroup {
	return g.Exclude(func(u Unit) bool {
		return u.Type() == t
	})
}

func (g UnitGroup) ExcludeNames(names ...UnitName) UnitGroup {
	return g.Exclude(func(u Unit) bool {
		for _, name := range names {
			if u.Name() == name {
				return true
			}
		}
		return false
	})
}

// MapUnitGroup performs a map operation on a UnitGroup
func MapUnitGroup[T any](g UnitGroup, p func(u Unit) T) []T {
	var mapped []T
	for _, u := range g {
		mapped = append(mapped, p(u))
	}

	return mapped
}
