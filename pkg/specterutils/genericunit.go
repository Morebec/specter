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

package specterutils

import (
	"fmt"
	"github.com/morebec/specter/pkg/specter"
	"github.com/zclconf/go-cty/cty"
)

// GenericUnit is a generic implementation of a Unit that saves its attributes in a list of attributes for introspection.
// these can be useful for loaders that are looser in what they allow.
type GenericUnit struct {
	UnitID     specter.UnitID
	typ        specter.UnitKind
	source     specter.Source
	Attributes []GenericUnitAttribute
}

func NewGenericUnit(name specter.UnitID, typ specter.UnitKind, source specter.Source) *GenericUnit {
	return &GenericUnit{UnitID: name, typ: typ, source: source}
}

func (u *GenericUnit) Description() string {
	if !u.HasAttribute("description") {
		return ""
	}

	attr := u.Attribute("description")
	return attr.Value.String()
}

func (u *GenericUnit) ID() specter.UnitID {
	return u.UnitID
}

func (u *GenericUnit) Kind() specter.UnitKind {
	return u.typ
}

func (u *GenericUnit) Source() specter.Source {
	return u.source
}

// Attribute returns an attribute by its name or nil if it was not found.
func (u *GenericUnit) Attribute(name string) *GenericUnitAttribute {
	for _, a := range u.Attributes {
		if a.Name == name {
			return &a
		}
	}

	return nil
}

// HasAttribute indicates if a unit has a certain attribute or not.
func (u *GenericUnit) HasAttribute(name string) bool {
	for _, a := range u.Attributes {
		if a.Name == name {
			return true
		}
	}
	return false
}

// AttributeType represents the type of attribute
type AttributeType string

// GenericUnitAttribute represents an attribute of a unit.
// It relies on cty.Value to represent the loaded value.
type GenericUnitAttribute struct {
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
	Attributes []GenericUnitAttribute
}

func (o ObjectValue) String() string {
	return fmt.Sprintf("ObjectValue{Kind: %s, Attributes: %v}", o.Type, o.Attributes)
}
