package specter

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"github.com/zclconf/go-cty/cty"
)

type SpecificationType string

type SpecificationName string

// Specification is a general purpose data structure to represent a specification as loaded from a file regardless of the loader
// used.
// It is the responsibility of a Module to convert a specification to an appropriate data structure representing the intent of a
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

	// Dependencies returns a list of the Names of the specifications this one depends on.
	Dependencies() []SpecificationName
}

type SpecBase struct {
	typ  SpecificationType
	name SpecificationName
	desc string
	src  Source
}

func (c SpecBase) Name() SpecificationName {
	return c.name
}

func (c SpecBase) Type() SpecificationType {
	return c.typ
}

func (c SpecBase) Description() string {
	return c.desc
}

func (c SpecBase) Source() Source {
	return c.src
}

// GenericSpecification is a generic implementation of a Specification that saves its attributes in a list of attributes for introspection.
// these can be useful for loaders that are looser in what they allow.
type GenericSpecification struct {
	name         SpecificationName
	typ          SpecificationType
	source       Source
	dependencies []SpecificationName
	Attributes   []GenericSpecAttribute
}

func (s GenericSpecification) SetSource(src Source) {
	s.source = src
}

func (s GenericSpecification) Description() string {
	if s.HasAttribute("description") {
		attr := s.Attribute("description")
		gAttr, ok := attr.Value.(GenericValue)
		if ok {
			return gAttr.AsString()
		}
	}

	return ""
}

func NewGenericSpecification(name SpecificationName, typ SpecificationType, source Source, dependencies []SpecificationName) *GenericSpecification {
	return &GenericSpecification{name: name, typ: typ, source: source, dependencies: dependencies}
}

func (s GenericSpecification) Name() SpecificationName {
	return s.name
}

func (s GenericSpecification) Type() SpecificationType {
	return s.typ
}

func (s GenericSpecification) Source() Source {
	return s.source
}

func (s GenericSpecification) Dependencies() []SpecificationName {
	return s.dependencies
}

// Attribute returns an attribute by its FilePath or nil if it was not found.
func (s GenericSpecification) Attribute(name string) *GenericSpecAttribute {
	for _, a := range s.Attributes {
		if a.Name == name {
			return &a
		}
	}

	return nil
}

// HasAttribute indicates if a specification has a certain attribute or not.
func (s GenericSpecification) HasAttribute(name string) bool {
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

// GenericSpecAttribute represents an attribute of a specification.
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

// MapSpecGroup performs a map operation on a SpecificationGroup
func MapSpecGroup[T any](g SpecificationGroup, p func(s Specification) T) []T {
	var mapped []T
	for _, s := range g {
		mapped = append(mapped, p(s))
	}

	return mapped
}

func UnexpectedSpecTypeError(actual SpecificationType, expected SpecificationType) error {
	return errors.NewWithMessage(errors.InternalErrorCode, fmt.Sprintf("expected specification of type %s, got %s", expected, actual))
}
