package specter

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"github.com/zclconf/go-cty/cty"
)

type SpecType string

type SpecName string

// Spec is a general purpose data structure to represent a spec as loaded from a file regardless of the loader
// used.
// It is the responsibility of a Module to convert a spec to an appropriate data structure representing the intent of a
// given Spec.
type Spec interface {
	// Name returns the unique FilePath of this spec.
	Name() SpecName

	// Type returns the type of this spec.
	Type() SpecType

	// Description of this spec.
	Description() string

	// Source returns the source of this spec.
	Source() Source

	// SetSource sets the source of the spec.
	// This method should only be used by loaders.
	SetSource(s Source)

	// Dependencies returns a list of the Names of the specs this one depends on.
	Dependencies() []SpecName
}

type SpecBase struct {
	typ  SpecType
	name SpecName
	desc string
	src  Source
}

func (c SpecBase) Name() SpecName {
	return c.name
}

func (c SpecBase) Type() SpecType {
	return c.typ
}

func (c SpecBase) Description() string {
	return c.desc
}

func (c SpecBase) Source() Source {
	return c.src
}

// GenericSpec is a generic implementation of a Spec that saves its attributes in a list of attributes for introspection.
// these can be useful for loaders that are looser in what they allow.
type GenericSpec struct {
	name         SpecName
	typ          SpecType
	source       Source
	dependencies []SpecName
	Attributes   []GenericSpecAttribute
}

func (s GenericSpec) SetSource(src Source) {
	s.source = src
}

func (s GenericSpec) Description() string {
	if s.HasAttribute("description") {
		attr := s.Attribute("description")
		gAttr, ok := attr.Value.(GenericValue)
		if ok {
			return gAttr.AsString()
		}
	}

	return ""
}

func NewGenericSpec(name SpecName, typ SpecType, source Source, dependencies []SpecName) *GenericSpec {
	return &GenericSpec{name: name, typ: typ, source: source, dependencies: dependencies}
}

func (s GenericSpec) Name() SpecName {
	return s.name
}

func (s GenericSpec) Type() SpecType {
	return s.typ
}

func (s GenericSpec) Source() Source {
	return s.source
}

func (s GenericSpec) Dependencies() []SpecName {
	return s.dependencies
}

// Attribute returns an attribute by its FilePath or nil if it was not found.
func (s GenericSpec) Attribute(name string) *GenericSpecAttribute {
	for _, a := range s.Attributes {
		if a.Name == name {
			return &a
		}
	}

	return nil
}

// HasAttribute indicates if a spec has a certain attribute or not.
func (s GenericSpec) HasAttribute(name string) bool {
	for _, a := range s.Attributes {
		if a.Name == name {
			return true
		}
	}
	return false
}

// AttributeType represents the type of an attribute
type AttributeType string

const (
	// Unknown is used for attributes where the actual type is unknown.
	Unknown = "any"
)

// GenericSpecAttribute represents an attribute of a spec.
// It relies on cty.Value to represent the loaded value.
type GenericSpecAttribute struct {
	Name  string
	Value AttributeValue
}

func (a GenericSpecAttribute) AsGenericValue() GenericValue {
	return a.Value.(GenericValue)
}

func (a GenericSpecAttribute) AsObjectValue() ObjectValue {
	return a.Value.(ObjectValue)
}

type AttributeValue interface {
	IsAttributeValue()
}

// GenericValue represents a generic value that is mostly unknown in terms of type and intent.
type GenericValue struct {
	cty.Value
}

func (d GenericValue) IsAttributeValue() {}

// ObjectValue represents a type of attribute value that is a nested data structure as opposed to a scalar value.
type ObjectValue struct {
	Type       AttributeType
	Attributes []GenericSpecAttribute
}

func (o ObjectValue) IsAttributeValue() {}

// SpecGroup Represents a list of Spec.
type SpecGroup []Spec

func NewSpecGroup(s ...Spec) SpecGroup {
	g := SpecGroup{}
	return append(g, s...)
}

// Merge Allows merging a group with another one.
func (g SpecGroup) Merge(group SpecGroup) SpecGroup {
	merged := g
	typeNameIndex := map[SpecName]any{}
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
func (g SpecGroup) Select(p func(s Spec) bool) SpecGroup {
	r := SpecGroup{}
	for _, s := range g {
		if p(s) {
			r = append(r, s)
		}
	}

	return r
}

func (g SpecGroup) SelectType(t SpecType) SpecGroup {
	return g.Select(func(s Spec) bool {
		return s.Type() == t
	})
}

func (g SpecGroup) SelectName(t SpecName) Spec {
	for _, s := range g {
		if s.Name() == t {
			return s
		}
	}

	return nil
}

func (g SpecGroup) Exclude(p func(s Spec) bool) SpecGroup {
	r := SpecGroup{}
	for _, s := range g {
		if !p(s) {
			r = append(r, s)
		}
	}

	return r
}

func (g SpecGroup) ExcludeType(t SpecType) SpecGroup {
	return g.Exclude(func(s Spec) bool {
		return s.Type() == t
	})
}

// MapSpecGroup performs a map operation on a SpecGroup
func MapSpecGroup[T any](g SpecGroup, p func(s Spec) T) []T {
	var mapped []T
	for _, s := range g {
		mapped = append(mapped, p(s))
	}

	return mapped
}

func UnexpectedSpecTypeError(actual SpecType, expected SpecType) error {
	return errors.NewWithMessage(errors.InternalErrorCode, fmt.Sprintf("expected spec of type %s, got %s", expected, actual))
}
