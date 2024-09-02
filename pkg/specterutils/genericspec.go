package specterutils

import (
	"fmt"
	"github.com/morebec/specter/pkg/specter"
	"github.com/zclconf/go-cty/cty"
)

// GenericSpecification is a generic implementation of a Specification that saves its attributes in a list of attributes for introspection.
// these can be useful for loaders that are looser in what they allow.
type GenericSpecification struct {
	name       specter.SpecificationName
	typ        specter.SpecificationType
	source     specter.Source
	Attributes []GenericSpecAttribute
}

func NewGenericSpecification(name specter.SpecificationName, typ specter.SpecificationType, source specter.Source) *GenericSpecification {
	return &GenericSpecification{name: name, typ: typ, source: source}
}

func (s *GenericSpecification) SetSource(src specter.Source) {
	s.source = src
}

func (s *GenericSpecification) Description() string {
	if !s.HasAttribute("description") {
		return ""
	}

	attr := s.Attribute("description")
	return attr.Value.String()
}

func (s *GenericSpecification) Name() specter.SpecificationName {
	return s.name
}

func (s *GenericSpecification) Type() specter.SpecificationType {
	return s.typ
}

func (s *GenericSpecification) Source() specter.Source {
	return s.source
}

// Attribute returns an attribute by its FilePath or nil if it was not found.
func (s *GenericSpecification) Attribute(name string) *GenericSpecAttribute {
	for _, a := range s.Attributes {
		if a.Name == name {
			return &a
		}
	}

	return nil
}

// HasAttribute indicates if a specification has a certain attribute or not.
func (s *GenericSpecification) HasAttribute(name string) bool {
	for _, a := range s.Attributes {
		if a.Name == name {
			return true
		}
	}
	return false
}

// AttributeType represents the type of attribute
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

type AttributeValue interface {
	String() string
}

var _ AttributeValue = GenericValue{}

// GenericValue represents a generic value that is mostly unknown in terms of type and intent.
type GenericValue struct {
	cty.Value
}

func (d GenericValue) String() string {
	switch d.Type() {
	case cty.String:
		return d.Value.AsString()
	default:
		return ""
	}
}

var _ AttributeValue = ObjectValue{}

// ObjectValue represents a type of attribute value that is a nested data structure as opposed to a scalar value.
type ObjectValue struct {
	Type       AttributeType
	Attributes []GenericSpecAttribute
}

func (o ObjectValue) String() string {
	return fmt.Sprintf("ObjectValue{Type: %s, Attributes: %v}", o.Type, o.Attributes)
}
