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

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/morebec/go-errors/errors"
	"github.com/zclconf/go-cty/cty"
)

const (
	HCLSourceFormat SourceFormat = "hcl"
)

const InvalidHCLErrorCode = "invalid_hcl"

// HCLGenericSpecLoader this SpecificationLoader loads Specifications as GenericSpecification.
type HCLGenericSpecLoader struct {
	hclparse.Parser
}

// NewHCLGenericSpecLoader this  SpecificationLoader will load all Specifications to instances of GenericSpecification.
func NewHCLGenericSpecLoader() *HCLGenericSpecLoader {
	return &HCLGenericSpecLoader{
		Parser: *hclparse.NewParser(),
	}
}

func (l HCLGenericSpecLoader) SupportsSource(s Source) bool {
	return s.Format == HCLSourceFormat
}

func (l HCLGenericSpecLoader) Load(s Source) ([]Specification, error) {
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Although the caller is responsible for calling HCLGenericSpecLoader.SupportsSource, guard against it.
	if !l.SupportsSource(s) {
		return nil, errors.NewWithMessage(
			UnsupportedSourceErrorCode,
			fmt.Sprintf(
				"invalid specification source %q, unsupported format %q",
				s.Location,
				s.Format,
			),
		)
	}

	file, diags := l.ParseHCL(s.Data, s.Location)
	if diags != nil && diags.HasErrors() {
		return nil, errors.Wrap(diags, InvalidHCLErrorCode)
	}

	var specifications []Specification

	body := file.Body.(*hclsyntax.Body)
	for _, block := range body.Blocks {
		// Ensure there is at least one label for the FilePath
		if len(block.Labels) == 0 || block.Labels[0] == "" {
			return nil, errors.NewWithMessage(
				InvalidHCLErrorCode,
				fmt.Sprintf(
					"invalid specification source %q at line %d:%d, block %q should contain a name",
					s.Location,
					block.Range().Start.Line,
					block.Range().Start.Column,
					block.Type,
				),
			)
		}

		// Extract Attributes in block.
		specAttributes, err := l.extractAttributesFromBlock(ctx, block)
		if err != nil {
			return nil, errors.WrapWithMessage(
				err,
				InvalidHCLErrorCode,
				fmt.Sprintf(
					"invalid specification source %q at line %d:%d for block %q",
					s.Location,
					block.Range().Start.Line,
					block.Range().Start.Column,
					block.Type,
				),
			)
		}

		// Create specification and add to list
		specifications = append(specifications, &GenericSpecification{
			name:       SpecificationName(block.Labels[0]),
			typ:        SpecificationType(block.Type),
			source:     s,
			Attributes: specAttributes,
		})
	}

	group := errors.NewGroup(errors.InternalErrorCode)
	if diags != nil && !diags.HasErrors() {
		for _, d := range diags.Errs() {
			group = group.Append(d)
		}
	}

	return specifications, errors.GroupOrNil(group)
}

func (l HCLGenericSpecLoader) extractAttributesFromBlock(ctx *hcl.EvalContext, block *hclsyntax.Block) ([]GenericSpecAttribute, hcl.Diagnostics) {
	var attrs []GenericSpecAttribute

	var diags hcl.Diagnostics

	// Detect attributes in current block.
	for _, a := range block.Body.Attributes {
		value, d := a.Expr.Value(ctx)
		if d != nil && d.HasErrors() {
			diags = append(diags, d...)
			continue
		}

		attrs = append(attrs, GenericSpecAttribute{
			Name:  a.Name,
			Value: GenericValue{value},
		})
	}

	// Handle nested blocks as attributes
	for _, b := range block.Body.Blocks {
		bName := ""
		if len(b.Labels) != 0 {
			bName = b.Labels[0]
		}

		bAttrs, d := l.extractAttributesFromBlock(ctx, b)
		if d.HasErrors() {
			diags = append(diags, d...)
			continue
		}

		attrs = append(attrs, GenericSpecAttribute{
			Name: bName,
			Value: ObjectValue{
				Type:       AttributeType(b.Type),
				Attributes: bAttrs,
			},
		})
	}

	return attrs, nil
}

type HCLSpecLoaderFileConfigurationProvider func() HCLFileConfig

type HCLFileConfig interface {
	Specifications() []Specification
}

// HCLVariableConfig represents a block configuration that allows defining variables.
type HCLVariableConfig struct {
	Name        string    `hcl:"FilePath,label"`
	Description string    `hcl:"description,optional"`
	Value       cty.Value `hcl:"value"`
}

// HCLSpecLoader this loader allows to load Specifications to typed structs by providing a HCLFileConfig.
type HCLSpecLoader struct {
	// represents the structure of a file that this HCL loader should support.
	parser             *hclparse.Parser
	fileConfigProvider HCLSpecLoaderFileConfigurationProvider
	evalCtx            *hcl.EvalContext
}

func NewHCLFileConfigSpecLoader(fileConfigProvider HCLSpecLoaderFileConfigurationProvider) *HCLSpecLoader {
	return &HCLSpecLoader{
		fileConfigProvider: fileConfigProvider,
		evalCtx: &hcl.EvalContext{
			Variables: map[string]cty.Value{},
		},
		parser: hclparse.NewParser(),
	}
}

func (l HCLSpecLoader) Load(s Source) ([]Specification, error) {
	ctx := l.evalCtx

	// Although the caller is responsible for calling HCLGenericSpecLoader.SupportsSource, guard against it.
	if !l.SupportsSource(s) {
		return nil, errors.NewWithMessage(
			UnsupportedSourceErrorCode,
			fmt.Sprintf(
				"invalid specification source %q, unsupported format %q",
				s.Location,
				s.Format,
			),
		)
	}

	// Parse const blocks to add them as Variables in the context.
	var diags hcl.Diagnostics
	var parsedFile *hcl.File
	parsedFile, diags = l.parser.ParseHCL(s.Data, s.Location)

	body := parsedFile.Body.(*hclsyntax.Body)
	for _, b := range body.Blocks {
		if b.Type == "const" {
			v, d := b.Body.Attributes["value"].Expr.Value(ctx)
			if d != nil && d.HasErrors() {
				diags = append(diags, d...)
			} else {
				ctx.Variables[b.Labels[0]] = v
			}
		}
	}

	// Decode config file
	fileConf := l.fileConfigProvider()
	err := hclsimple.Decode(s.Location, s.Data, ctx, fileConf)

	if err != nil {
		var d hcl.Diagnostics
		if !errors.As(err, &d) {
			return nil, err
		}
		diags = append(diags, d...)
	}

	if diags != nil && diags.HasErrors() {
		return nil, diags
	}

	// Set source for all specifications
	specifications := fileConf.Specifications()
	for _, sp := range specifications {
		sp.SetSource(s)
	}
	return specifications, nil
}

func (l HCLSpecLoader) SupportsSource(s Source) bool {
	return s.Format == HCLSourceFormat
}
