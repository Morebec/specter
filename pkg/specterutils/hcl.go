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
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/morebec/go-errors/errors"
	"github.com/morebec/specter/pkg/specter"
	"github.com/zclconf/go-cty/cty"
)

const (
	HCLSourceFormat specter.SourceFormat = "hcl"
)

const InvalidHCLErrorCode = "specter.spec_loading.invalid_hcl"

// NewHCLGenericSpecLoader this  SpecificationLoader will load all Specifications to instances of GenericSpecification.
func NewHCLGenericSpecLoader() *HCLGenericSpecLoader {
	return &HCLGenericSpecLoader{
		Parser: *hclparse.NewParser(),
	}
}

// HCLGenericSpecLoader this SpecificationLoader loads Specifications as GenericSpecification.
type HCLGenericSpecLoader struct {
	hclparse.Parser
}

func (l HCLGenericSpecLoader) SupportsSource(s specter.Source) bool {
	return s.Format == HCLSourceFormat
}

func (l HCLGenericSpecLoader) Load(s specter.Source) ([]specter.Specification, error) {
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Although the caller is responsible for calling HCLGenericSpecLoader.SupportsSource, guard against it.
	if !l.SupportsSource(s) {
		return nil, errors.NewWithMessage(
			specter.UnsupportedSourceErrorCode,
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

	var specifications []specter.Specification

	body := file.Body.(*hclsyntax.Body)
	for _, block := range body.Blocks {
		// Ensure there is at least one label for the block
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
		specAttributes, attrDiags := l.extractAttributesFromBlock(ctx, block)
		if attrDiags != nil {
			diags = diags.Extend(attrDiags)
			continue
			//return nil, errors.WrapWithMessage(
			//	err,
			//	InvalidHCLErrorCode,
			//	fmt.Sprintf(
			//		"invalid specification source %q at line %d:%d for block %q",
			//		s.Location,
			//		block.Range().Start.Line,
			//		block.Range().Start.Column,
			//		block.Type,
			//	),
			//)
		}

		// Create specification and add to list
		specifications = append(specifications, &GenericSpecification{
			name:       specter.SpecificationName(block.Labels[0]),
			typ:        specter.SpecificationType(block.Type),
			source:     s,
			Attributes: specAttributes,
		})
	}

	group := errors.NewGroup(InvalidHCLErrorCode)
	if diags != nil && diags.HasErrors() {
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

	return attrs, diags
}

type HCLSpecLoaderFileConfigurationProvider func() HCLFileConfig

// HCLFileConfig interface that is to be implemented to define the structure of HCL specification files.
type HCLFileConfig interface {
	Specifications() []specter.Specification
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

func NewHCLSpecLoader(fileConfigProvider HCLSpecLoaderFileConfigurationProvider) *HCLSpecLoader {
	return &HCLSpecLoader{
		fileConfigProvider: fileConfigProvider,
		evalCtx: &hcl.EvalContext{
			Variables: map[string]cty.Value{},
		},
		parser: hclparse.NewParser(),
	}
}

func (l HCLSpecLoader) Load(s specter.Source) ([]specter.Specification, error) {
	// Although the caller is responsible for calling HCLGenericSpecLoader.SupportsSource, guard against it.
	if !l.SupportsSource(s) {
		return nil, errors.NewWithMessage(
			specter.UnsupportedSourceErrorCode,
			fmt.Sprintf(
				"invalid specification source %q, unsupported format %q",
				s.Location,
				s.Format,
			),
		)
	}

	ctx := l.evalCtx
	//// Parse const blocks to add them as Variables in the context.
	//var diags hcl.Diagnostics
	//var parsedFile *hcl.File
	//parsedFile, diags = l.parser.ParseHCL(s.Data, s.Location)
	//
	//body := parsedFile.Body.(*hclsyntax.Body)
	//for _, b := range body.Blocks {
	//	if b.Type == "const" {
	//		v, d := b.Body.Attributes["value"].Expr.Value(ctx)
	//		if d != nil && d.HasErrors() {
	//			diags = append(diags, d...)
	//		} else {
	//			ctx.Variables[b.Labels[0]] = v
	//		}
	//	}
	//}

	if len(s.Data) == 0 {
		return nil, nil
	}

	// Decode config file
	fileConf := l.fileConfigProvider()
	err := hclsimple.Decode(s.Location, s.Data, ctx, fileConf)

	if err != nil {
		return nil, errors.Wrap(err, InvalidHCLErrorCode)
	}

	// Set source for all specifications
	specifications := fileConf.Specifications()
	for _, sp := range specifications {
		sp.SetSource(s)
	}
	return specifications, nil
}

func (l HCLSpecLoader) SupportsSource(s specter.Source) bool {
	return s.Format == HCLSourceFormat
}
