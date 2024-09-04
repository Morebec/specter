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

package specterutils_test

import (
	"github.com/morebec/specter/pkg/specter"
	"github.com/morebec/specter/pkg/specterutils"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
	"testing"
)

func TestGenericUnit_Description(t *testing.T) {
	tests := []struct {
		name  string
		given *specterutils.GenericUnit
		then  string
	}{
		{
			name: "GIVEN a unit with a description attribute THEN return the description",
			given: &specterutils.GenericUnit{
				Attributes: []specterutils.GenericUnitAttribute{
					{
						Name:  "description",
						Value: specterutils.GenericValue{Value: cty.StringVal("This is a test unit")},
					},
				},
			},
			then: "This is a test unit",
		},
		{
			name: "GIVEN a unit without a description attribute THEN return an empty string",
			given: &specterutils.GenericUnit{
				Attributes: []specterutils.GenericUnitAttribute{},
			},
			then: "",
		},
		{
			name: "GIVEN a unit with a non-string description THEN return an empty string",
			given: &specterutils.GenericUnit{
				Attributes: []specterutils.GenericUnitAttribute{
					{
						Name:  "description",
						Value: specterutils.GenericValue{Value: cty.NumberIntVal(42)}, // Not a string value
					},
				},
			},
			then: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.Description()
			assert.Equal(t, tt.then, got)
		})
	}
}

func TestGenericUnit_Attribute(t *testing.T) {
	tests := []struct {
		name  string
		given *specterutils.GenericUnit
		when  string
		then  *specterutils.GenericUnitAttribute
	}{
		{
			name: "GIVEN a unit with a specific attribute WHEN Attribute is called THEN return the attribute",
			given: &specterutils.GenericUnit{
				Attributes: []specterutils.GenericUnitAttribute{
					{
						Name:  "attr1",
						Value: specterutils.GenericValue{Value: cty.StringVal("value1")},
					},
				},
			},
			when: "attr1",
			then: &specterutils.GenericUnitAttribute{
				Name:  "attr1",
				Value: specterutils.GenericValue{Value: cty.StringVal("value1")},
			},
		},
		{
			name: "GIVEN a unit without the specified attribute WHEN Attribute is called THEN return nil",
			given: &specterutils.GenericUnit{
				Attributes: []specterutils.GenericUnitAttribute{},
			},
			when: "nonexistent",
			then: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.Attribute(tt.when)
			assert.Equal(t, tt.then, got)
		})
	}
}

func TestGenericUnit_HasAttribute(t *testing.T) {
	tests := []struct {
		name  string
		given *specterutils.GenericUnit
		when  string
		then  bool
	}{
		{
			name: "GIVEN a unit with a specific attribute WHEN HasAttribute is called THEN return true",
			given: &specterutils.GenericUnit{
				Attributes: []specterutils.GenericUnitAttribute{
					{
						Name:  "attr1",
						Value: specterutils.GenericValue{Value: cty.StringVal("value1")},
					},
				},
			},
			when: "attr1",
			then: true,
		},
		{
			name: "GIVEN a unit without the specified attribute WHEN HasAttribute is called THEN return false",
			given: &specterutils.GenericUnit{
				Attributes: []specterutils.GenericUnitAttribute{},
			},
			when: "nonexistent",
			then: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.given.HasAttribute(tt.when)
			assert.Equal(t, tt.then, got)
		})
	}
}

func TestGenericUnit_SetSource(t *testing.T) {
	tests := []struct {
		name  string
		given *specterutils.GenericUnit
		when  specter.Source
		then  specter.Source
	}{
		{
			name:  "GIVEN a unit WHEN SetSource is called THEN updates the source",
			given: specterutils.NewGenericUnit("name", "type", specter.Source{Location: "initial/path"}),
			when:  specter.Source{Location: "new/path"},
			then:  specter.Source{Location: "new/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.given.SetSource(tt.when)
			assert.Equal(t, tt.then, tt.given.Source())
		})
	}
}

func TestObjectValue_String(t *testing.T) {
	o := specterutils.ObjectValue{Type: "hello", Attributes: []specterutils.GenericUnitAttribute{
		{Name: "hello", Value: specterutils.GenericValue{Value: cty.StringVal("world")}},
	}}
	assert.Equal(t, "ObjectValue{Type: hello, Attributes: [{hello world}]}", o.String())
}
