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

func TestGenericSpecification_Description(t *testing.T) {
	tests := []struct {
		name  string
		given *specterutils.GenericSpecification
		then  string
	}{
		{
			name: "GIVEN a specification with a description attribute THEN return the description",
			given: &specterutils.GenericSpecification{
				Attributes: []specterutils.GenericSpecAttribute{
					{
						Name:  "description",
						Value: specterutils.GenericValue{Value: cty.StringVal("This is a test specification")},
					},
				},
			},
			then: "This is a test specification",
		},
		{
			name: "GIVEN a specification without a description attribute THEN return an empty string",
			given: &specterutils.GenericSpecification{
				Attributes: []specterutils.GenericSpecAttribute{},
			},
			then: "",
		},
		{
			name: "GIVEN a specification with a non-string description THEN return an empty string",
			given: &specterutils.GenericSpecification{
				Attributes: []specterutils.GenericSpecAttribute{
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

func TestGenericSpecification_Attribute(t *testing.T) {
	tests := []struct {
		name  string
		given *specterutils.GenericSpecification
		when  string
		then  *specterutils.GenericSpecAttribute
	}{
		{
			name: "GIVEN a specification with a specific attribute WHEN Attribute is called THEN return the attribute",
			given: &specterutils.GenericSpecification{
				Attributes: []specterutils.GenericSpecAttribute{
					{
						Name:  "attr1",
						Value: specterutils.GenericValue{Value: cty.StringVal("value1")},
					},
				},
			},
			when: "attr1",
			then: &specterutils.GenericSpecAttribute{
				Name:  "attr1",
				Value: specterutils.GenericValue{Value: cty.StringVal("value1")},
			},
		},
		{
			name: "GIVEN a specification without the specified attribute WHEN Attribute is called THEN return nil",
			given: &specterutils.GenericSpecification{
				Attributes: []specterutils.GenericSpecAttribute{},
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

func TestGenericSpecification_HasAttribute(t *testing.T) {
	tests := []struct {
		name  string
		given *specterutils.GenericSpecification
		when  string
		then  bool
	}{
		{
			name: "GIVEN a specification with a specific attribute WHEN HasAttribute is called THEN return true",
			given: &specterutils.GenericSpecification{
				Attributes: []specterutils.GenericSpecAttribute{
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
			name: "GIVEN a specification without the specified attribute WHEN HasAttribute is called THEN return false",
			given: &specterutils.GenericSpecification{
				Attributes: []specterutils.GenericSpecAttribute{},
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

func TestGenericSpecification_SetSource(t *testing.T) {
	tests := []struct {
		name  string
		given *specterutils.GenericSpecification
		when  specter.Source
		then  specter.Source
	}{
		{
			name:  "GIVEN a specification WHEN SetSource is called THEN updates the source",
			given: specterutils.NewGenericSpecification("name", "type", specter.Source{Location: "initial/path"}),
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
	o := specterutils.ObjectValue{Type: "hello", Attributes: []specterutils.GenericSpecAttribute{
		{Name: "hello", Value: specterutils.GenericValue{Value: cty.StringVal("world")}},
	}}
	assert.Equal(t, "ObjectValue{Type: hello, Attributes: [{hello world}]}", o.String())
}
