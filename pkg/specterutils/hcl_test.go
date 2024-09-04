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
	"github.com/morebec/specter/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"testing"
)

func TestHCLGenericUnitLoader_SupportsSource(t *testing.T) {
	type when struct {
		source specter.Source
	}
	type then struct {
		supports bool
	}
	tests := []struct {
		name string
		when when
		then then
	}{
		{
			name: "WHEN a non HCL format THEN return false",
			when: when{
				specter.Source{Format: specterutils.HCLSourceFormat},
			},
			then: then{supports: true},
		},
		{
			name: "WHEN a non HCL format THEN return false",
			when: when{
				specter.Source{Format: "txt"},
			},
			then: then{supports: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := specterutils.NewHCLGenericUnitLoader()
			assert.Equalf(t, tt.then.supports, l.SupportsSource(tt.when.source), "SupportsSource(%v)", tt.when.source)
		})
	}
}

func TestHCLGenericUnitLoader_Load(t *testing.T) {
	type when struct {
		source specter.Source
	}
	type then struct {
		expectedUnits []specter.Unit
		expectedError require.ErrorAssertionFunc
	}

	mockFile := HclConfigMock{}

	tests := []struct {
		name string
		when when
		then then
	}{
		{
			name: "WHEN an empty file THEN return nil",
			when: when{
				source: specter.Source{
					Format: specterutils.HCLSourceFormat,
					Data:   []byte(``),
				},
			},
			then: then{
				expectedUnits: nil,
				expectedError: require.NoError,
			},
		},
		{
			name: "WHEN a valid hcl file THEN the specs should be returned an no error",
			when: when{
				source: mockFile.source(),
			},
			then: then{
				expectedUnits: []specter.Unit{
					mockFile.genericUnit(),
				},
				expectedError: require.NoError,
			},
		},
		{
			name: "WHEN an unsupported file format THEN an error should be returned",
			when: when{
				source: specter.Source{
					Format: "txt",
				},
			},
			then: then{
				expectedUnits: nil,
				expectedError: testutils.RequireErrorWithCode(specter.UnsupportedSourceErrorCode),
			},
		},
		{
			name: "WHEN an unparsable hcl file THEN an error should be returned",
			when: when{
				source: specter.Source{
					Data: []byte(`
con st = var o
`),
					Format: specterutils.HCLSourceFormat,
				},
			},
			then: then{
				expectedUnits: nil,
				expectedError: testutils.RequireErrorWithCode(specterutils.InvalidHCLErrorCode),
			},
		},
		{
			name: "WHEN a unit type without name THEN an error should be returned",
			when: when{
				source: specter.Source{
					Data: []byte(`
block {
}
`),
					Format: specterutils.HCLSourceFormat,
				},
			},
			then: then{
				expectedUnits: nil,
				expectedError: testutils.RequireErrorWithCode(specterutils.InvalidHCLErrorCode),
			},
		},
		// ATTRIBUTES
		{
			name: "WHEN an attribute is invalid THEN an error should be returned",
			when: when{
				source: specter.Source{
					Data: []byte(`
specType "specName" {
	attribute = var.example ? 12 : "hello"
}
`),
					Format: specterutils.HCLSourceFormat,
				},
			},
			then: then{
				expectedUnits: nil,
				expectedError: testutils.RequireErrorWithCode(specterutils.InvalidHCLErrorCode),
			},
		},
		{
			name: "WHEN an attribute in a nested block is invalid THEN an error should be returned",
			when: when{
				source: specter.Source{
					Data: []byte(`
specType "specName" {
	block "name" {
		attribute = var.example ? 12 : "hello"
	}
}
`),
					Format: specterutils.HCLSourceFormat,
				},
			},
			then: then{
				expectedUnits: nil,
				expectedError: testutils.RequireErrorWithCode(specterutils.InvalidHCLErrorCode),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := specterutils.NewHCLGenericUnitLoader()

			actualUnits, err := l.Load(tt.when.source)
			if tt.then.expectedError != nil {
				tt.then.expectedError(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.then.expectedUnits, actualUnits)
		})
	}
}

// CUSTOM CONFIG FILES //

func TestHCLUnitLoader_Load(t *testing.T) {
	type when struct {
		source specter.Source
	}
	type then struct {
		expectedUnits []specter.Unit
		expectedError require.ErrorAssertionFunc
	}

	mockFile := HclConfigMock{}

	tests := []struct {
		name string
		when when
		then then
	}{
		{
			name: "WHEN an empty file THEN return nil",
			when: when{
				source: specter.Source{
					Format: specterutils.HCLSourceFormat,
					Data:   []byte(``),
				},
			},
			then: then{
				expectedUnits: nil,
				expectedError: require.NoError,
			},
		},
		{
			name: "WHEN an unsupported file format THEN an error should be returned",
			when: when{
				source: specter.Source{
					Format: "txt",
				},
			},
			then: then{
				expectedUnits: nil,
				expectedError: testutils.RequireErrorWithCode(specter.UnsupportedSourceErrorCode),
			},
		},
		{
			name: "WHEN an unparsable hcl file THEN an error should be returned",
			when: when{
				source: specter.Source{
					Data: []byte(`
con st = var o
`),
					Format: specterutils.HCLSourceFormat,
				},
			},
			then: then{
				expectedUnits: nil,
				expectedError: testutils.RequireErrorWithCode(specterutils.InvalidHCLErrorCode),
			},
		},
		{
			name: "WHEN valid hcl file THEN return units",
			when: when{
				source: mockFile.source(),
			},
			then: then{
				expectedUnits: []specter.Unit{
					mockFile.genericUnit(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := specterutils.NewHCLUnitLoader(func() specterutils.HCLFileConfig {
				return &HclConfigMock{}
			})

			actualUnits, err := l.Load(tt.when.source)
			if tt.then.expectedError != nil {
				tt.then.expectedError(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.then.expectedUnits, actualUnits)
		})
	}
}

var _ specterutils.HCLFileConfig = (*HclConfigMock)(nil)

type HclConfigMock struct {
	Service struct {
		Name        string `hcl:"name,label"`
		Image       string `hcl:"image"`
		Environment struct {
			Name              string `hcl:"name,label"`
			MySQLRootPassword string `hcl:"MYSQL_ROOT_PASSWORD"`
		} `hcl:"environment,block"`
	} `hcl:"service,block"`
}

func (m *HclConfigMock) data() []byte {
	return []byte(`
service "specter" {
	image = "specter:1.0.0"
	environment "dev" {
		MYSQL_ROOT_PASSWORD = "password"
	}
}
`)
}

func (m *HclConfigMock) Units() []specter.Unit {
	return []specter.Unit{
		m.genericUnit(),
	}
}

func (m *HclConfigMock) genericUnit() *specterutils.GenericUnit {
	unit := specterutils.NewGenericUnit("specter", "service", m.source())

	unit.Attributes = append(unit.Attributes, specterutils.GenericUnitAttribute{
		Name: "image",
		Value: specterutils.GenericValue{
			Value: cty.StringVal("specter:1.0.0"),
		},
	})
	unit.Attributes = append(unit.Attributes, specterutils.GenericUnitAttribute{
		Name: "dev",
		Value: specterutils.ObjectValue{
			Type: "environment",
			Attributes: []specterutils.GenericUnitAttribute{
				{
					Name: "MYSQL_ROOT_PASSWORD",
					Value: specterutils.GenericValue{
						Value: cty.StringVal("password"),
					},
				},
			},
		},
	})
	return unit
}

func (m *HclConfigMock) source() specter.Source {
	return specter.Source{
		Location: "specter.hcl",
		Data:     m.data(),
		Format:   specterutils.HCLSourceFormat,
	}
}
