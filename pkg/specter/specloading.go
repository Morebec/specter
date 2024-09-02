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

// UnsupportedSourceErrorCode ErrorSeverity code returned by a SpecificationLoader when a given loader does not support a certain source.
const UnsupportedSourceErrorCode = "specter.spec_loading.unsupported_source"

// SpecificationLoader is a service responsible for loading Specifications from Sources.
type SpecificationLoader interface {
	// Load loads a slice of Specification from a Source, or returns an error if it encountered a failure.
	Load(s Source) ([]Specification, error)

	// SupportsSource indicates if this loader supports a certain source or not.
	SupportsSource(s Source) bool
}

type SpecificationType string

type SpecificationName string

// Specification is a general purpose data structure to represent a specification as loaded from a file regardless of the loader
// used.
// It is the responsibility of the application using specter to convert a specification to an appropriate data structure representing the intent of a
// given Specification.
type Specification interface {
	// Name returns the unique Name of this specification.
	Name() SpecificationName

	// Type returns the type of this specification.
	Type() SpecificationType

	// Description of this specification.
	Description() string

	// Source returns the source of this specification.
	Source() Source

	// SetSource sets the source of the specification.
	// This method should only be used by loaders.
	SetSource(s Source)
}

// SpecificationGroup Represents a list of Specification.
type SpecificationGroup []Specification

func NewSpecGroup(s ...Specification) SpecificationGroup {
	g := SpecificationGroup{}
	return append(g, s...)
}

// Merge Allows merging a group with another one.
func (g SpecificationGroup) Merge(group SpecificationGroup) SpecificationGroup {
	merged := g
	typeNameIndex := map[SpecificationName]any{}
	for _, s := range g {
		typeNameIndex[s.Name()] = nil
	}
	for _, s := range group {
		if _, found := typeNameIndex[s.Name()]; found {
			continue
		}
		typeNameIndex[s.Name()] = nil
		merged = append(merged, s)
	}
	return merged
}

// Select allows filtering the group for certain specifications.
func (g SpecificationGroup) Select(p func(s Specification) bool) SpecificationGroup {
	r := SpecificationGroup{}
	for _, s := range g {
		if p(s) {
			r = append(r, s)
		}
	}

	return r
}

func (g SpecificationGroup) SelectType(t SpecificationType) SpecificationGroup {
	return g.Select(func(s Specification) bool {
		return s.Type() == t
	})
}

func (g SpecificationGroup) SelectName(t SpecificationName) Specification {
	for _, s := range g {
		if s.Name() == t {
			return s
		}
	}

	return nil
}

func (g SpecificationGroup) SelectNames(names ...SpecificationName) SpecificationGroup {
	return g.Select(func(s Specification) bool {
		for _, name := range names {
			if s.Name() == name {
				return true
			}
		}
		return false
	})
}

func (g SpecificationGroup) Exclude(p func(s Specification) bool) SpecificationGroup {
	r := SpecificationGroup{}
	for _, s := range g {
		if !p(s) {
			r = append(r, s)
		}
	}

	return r
}

func (g SpecificationGroup) ExcludeType(t SpecificationType) SpecificationGroup {
	return g.Exclude(func(s Specification) bool {
		return s.Type() == t
	})
}

func (g SpecificationGroup) ExcludeNames(names ...SpecificationName) SpecificationGroup {
	return g.Exclude(func(s Specification) bool {
		for _, name := range names {
			if s.Name() == name {
				return true
			}
		}
		return false
	})
}

// MapSpecGroup performs a map operation on a SpecificationGroup
func MapSpecGroup[T any](g SpecificationGroup, p func(s Specification) T) []T {
	var mapped []T
	for _, s := range g {
		mapped = append(mapped, p(s))
	}

	return mapped
}
