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

type UnitKind string

type UnitID string

// Unit is a general purpose data structure to represent a unit as loaded from a file regardless of the loader
// used.
// It is the responsibility of the application using specter to convert a unit to an appropriate data structure representing the intent of a
// given Unit.
type Unit interface {
	// ID returns the unique Name of this unit.
	ID() UnitID

	// Kind returns the type of this unit.
	Kind() UnitKind

	// Source returns the source of this unit.
	Source() Source
}

// WrappingUnit is a generic implementation of a Unit that wraps an underlying value.
// This allows users to pass any value, even those that do not implement the Unit interface,
// through the Spectre pipeline. The wrapped value can be later unwrapped and used as needed.
//
// WrappingUnit provides a flexible solution for users who want to avoid directly
// implementing the Unit interface in their own types or who are dealing with
// primitive types or external structs.
//
// T represents the type of the value being wrapped.
//
// Example usage:
//
//	wrapped := Spectre.UnitOf(myValue)
//	unwrapped := wrapped.Unwrap()
type WrappingUnit[T any] struct {
	id      UnitID
	kind    UnitKind
	source  Source
	wrapped T
}

func UnitOf[T any](v T, id UnitID, kind UnitKind, source Source) *WrappingUnit[T] {
	return &WrappingUnit[T]{
		id:      id,
		kind:    kind,
		source:  source,
		wrapped: v,
	}
}

func (w *WrappingUnit[T]) ID() UnitID {
	return w.id
}

func (w *WrappingUnit[T]) Kind() UnitKind {
	return w.kind
}

func (w *WrappingUnit[T]) Source() Source {
	return w.source
}

// Unwrap returns the wrapped value.
func (w *WrappingUnit[T]) Unwrap() T {
	return w.wrapped
}

// UnitLoader is a service responsible for loading Units from Sources.
type UnitLoader interface {
	// Load loads a slice of Unit from a Source, or returns an error if it encountered a failure.
	Load(s Source) ([]Unit, error)

	// SupportsSource indicates if this loader supports a certain source or not.
	SupportsSource(s Source) bool
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
	idIndex := map[UnitID]any{}
	for _, u := range g {
		idIndex[u.ID()] = nil
	}
	for _, u := range group {
		if _, found := idIndex[u.ID()]; found {
			continue
		}
		idIndex[u.ID()] = nil
		merged = append(merged, u)
	}
	return merged
}

type UnitMatcher func(u Unit) bool

func UnitWithKindMatcher(kind UnitKind) UnitMatcher {
	return func(u Unit) bool {
		return u.Kind() == kind
	}
}

// Select allows filtering the group for certain units that match a certain UnitMatcher.
func (g UnitGroup) Select(m UnitMatcher) UnitGroup {
	r := UnitGroup{}
	for _, u := range g {
		if m(u) {
			r = append(r, u)
		}
	}

	return r
}

// Find search for a unit matching the given UnitMatcher.
func (g UnitGroup) Find(m UnitMatcher) (Unit, bool) {
	for _, u := range g {
		if m(u) {
			return u, true
		}
	}
	return nil, false
}

func (g UnitGroup) SelectType(t UnitKind) UnitGroup {
	return g.Select(func(u Unit) bool {
		return u.Kind() == t
	})
}

func (g UnitGroup) SelectName(t UnitID) Unit {
	for _, u := range g {
		if u.ID() == t {
			return u
		}
	}

	return nil
}

func (g UnitGroup) SelectNames(names ...UnitID) UnitGroup {
	return g.Select(func(u Unit) bool {
		for _, name := range names {
			if u.ID() == name {
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

func (g UnitGroup) ExcludeType(t UnitKind) UnitGroup {
	return g.Exclude(func(u Unit) bool {
		return u.Kind() == t
	})
}

func (g UnitGroup) ExcludeNames(names ...UnitID) UnitGroup {
	return g.Exclude(func(u Unit) bool {
		for _, name := range names {
			if u.ID() == name {
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
