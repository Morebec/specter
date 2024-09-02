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
	"github.com/morebec/go-errors/errors"
	"github.com/morebec/specter/pkg/specter"
	"strings"
	"unicode"
)

const LinterResultArtifactID = "_linting_processor_results"

// UndefinedSpecificationName constant used to test against undefined SpecificationName.
const UndefinedSpecificationName specter.SpecificationName = ""

const LintingErrorCode = "specter.spec_processing.linting_error"

type LinterResultSeverity string

const (
	ErrorSeverity   LinterResultSeverity = "error"
	WarningSeverity LinterResultSeverity = "warning"
)

type LintingProcessor struct {
	linters []SpecificationLinter
}

func NewLintingProcessor(linters ...SpecificationLinter) *LintingProcessor {
	return &LintingProcessor{linters: linters}
}

func (l LintingProcessor) Name() string {
	return "linting_processor"
}

func (l LintingProcessor) Process(ctx specter.ProcessingContext) (artifacts []specter.Artifact, err error) {
	linter := CompositeSpecificationLinter(l.linters...)
	ctx.Logger.Info("\nLinting specifications ...")

	lr := linter.Lint(ctx.Specifications)

	artifacts = append(artifacts, lr)

	if lr.HasWarnings() {
		for _, w := range lr.Warnings() {
			ctx.Logger.Warning(fmt.Sprintf("Warning: %s\n", w.Message))
		}
	}

	if lr.HasErrors() {
		for _, e := range lr.Errors().Errors {
			ctx.Logger.Error(fmt.Sprintf("Error: %s\n", e.Error()))
		}
		err = lr.Errors()
	}

	if !lr.HasWarnings() && !lr.HasErrors() {
		ctx.Logger.Success("Specifications linted successfully.")
	}

	return artifacts, err

}

func GetLintingResultsFromContext(ctx specter.ProcessingContext) LinterResultSet {
	return specter.GetContextArtifact[LinterResultSet](ctx, LinterResultArtifactID)
}

type LinterResult struct {
	Severity LinterResultSeverity
	Message  string
}

// LinterResultSet represents a set of LinterResult.
type LinterResultSet []LinterResult

// Errors returns a list of LinterResult as errors.
func (s LinterResultSet) Errors() errors.Group {
	errs := errors.NewGroup(LintingErrorCode)

	for _, r := range s {
		if r.Severity == ErrorSeverity {
			errs = errs.Append(
				errors.NewWithMessage(LintingErrorCode, r.Message),
			)
		}
	}

	return errs
}

// Warnings returns another LinterResultSet with only warnings.
func (s LinterResultSet) Warnings() LinterResultSet {
	warns := LinterResultSet{}

	for _, r := range s {
		if r.Severity == WarningSeverity {
			warns = append(warns, r)
		}
	}

	return warns
}

func (s LinterResultSet) ID() specter.ArtifactID {
	return LinterResultArtifactID
}

// HasErrors returns if this result set has any result representing an error.
func (s LinterResultSet) HasErrors() bool {
	return s.Errors().HasErrors()
}

// HasWarnings returns if this result set has any result representing a warning.
func (s LinterResultSet) HasWarnings() bool {
	return len(s.Warnings()) != 0
}

// SpecificationLinter represents a function responsible for linting specifications.
type SpecificationLinter interface {
	Lint(specifications specter.SpecificationGroup) LinterResultSet
}

// SpecificationLinterFunc implementation of a SpecificationLinter that relies on a func
type SpecificationLinterFunc func(specifications specter.SpecificationGroup) LinterResultSet

func (l SpecificationLinterFunc) Lint(specifications specter.SpecificationGroup) LinterResultSet {
	return l(specifications)
}

// CompositeSpecificationLinter A Composite linter is responsible for running multiple linters as one.
func CompositeSpecificationLinter(linters ...SpecificationLinter) SpecificationLinterFunc {
	return func(specifications specter.SpecificationGroup) LinterResultSet {
		var result LinterResultSet
		for _, linter := range linters {
			lr := linter.Lint(specifications)
			result = append(result, lr...)
		}
		return result
	}
}

// SpecificationMustNotHaveUndefinedNames ensures that no specification has an undefined name
func SpecificationMustNotHaveUndefinedNames(severity LinterResultSeverity) SpecificationLinterFunc {
	return func(specifications specter.SpecificationGroup) LinterResultSet {
		var result LinterResultSet

		for _, s := range specifications {
			if s.Name() == UndefinedSpecificationName {
				result = append(result, LinterResult{
					Severity: severity,
					Message:  fmt.Sprintf("specification at %q has an undefined name", s.Source().Location),
				})
			}
		}

		return result
	}
}

// SpecificationsMustHaveUniqueNames ensures that names are unique amongst specifications.
func SpecificationsMustHaveUniqueNames(severity LinterResultSeverity) SpecificationLinterFunc {
	return func(specifications specter.SpecificationGroup) LinterResultSet {

		var result LinterResultSet

		// Where key is the type FilePath and the array contains all the specification file locations where it was encountered.
		encounteredNames := map[specter.SpecificationName][]string{}

		for _, s := range specifications {
			if _, found := encounteredNames[s.Name()]; found {
				encounteredNames[s.Name()] = append(encounteredNames[s.Name()], s.Source().Location)
			} else {
				encounteredNames[s.Name()] = []string{s.Source().Location}
			}
		}

		for name, files := range encounteredNames {
			if len(files) > 1 {
				// Deduplicate
				fnMap := map[string]struct{}{}
				for _, fn := range files {
					fnMap[fn] = struct{}{}
				}
				var fileNames []string
				for fn := range fnMap {
					fileNames = append(fileNames, fn)
				}

				result = append(result, LinterResult{
					Severity: severity,
					Message: fmt.Sprintf(
						"duplicate specification name detected for %q in the following file(s): %s",
						name,
						strings.Join(fileNames, ", "),
					),
				})
			}
		}

		return result
	}
}

// SpecificationsMustHaveDescriptionAttribute ensures that all specifications have a description.
func SpecificationsMustHaveDescriptionAttribute(severity LinterResultSeverity) SpecificationLinterFunc {
	return func(specifications specter.SpecificationGroup) LinterResultSet {
		var result LinterResultSet
		for _, s := range specifications {
			if s.Description() == "" {
				result = append(result, LinterResult{
					Severity: severity,
					Message:  fmt.Sprintf("specification %q at location %q does not have a description", s.Name(), s.Source().Location),
				})
			}
		}
		return result
	}
}

// SpecificationsDescriptionsMustStartWithACapitalLetter ensures that specification descriptions start with a capital letter.
func SpecificationsDescriptionsMustStartWithACapitalLetter(severity LinterResultSeverity) SpecificationLinterFunc {
	return func(specifications specter.SpecificationGroup) LinterResultSet {
		var result LinterResultSet
		for _, s := range specifications {
			if s.Description() != "" {
				firstLetter := rune(s.Description()[0])
				if unicode.IsUpper(firstLetter) {
					continue
				}
			}
			result = append(result, LinterResult{
				Severity: severity,
				Message:  fmt.Sprintf("the description of specification %q at location %q does not start with a capital letter", s.Name(), s.Source().Location),
			})
		}
		return result
	}
}

// SpecificationsDescriptionsMustEndWithPeriod ensures that specification descriptions end with a period.
func SpecificationsDescriptionsMustEndWithPeriod(severity LinterResultSeverity) SpecificationLinterFunc {
	return func(specifications specter.SpecificationGroup) LinterResultSet {
		var result LinterResultSet
		for _, s := range specifications {
			if !strings.HasSuffix(s.Description(), ".") {
				result = append(result, LinterResult{
					Severity: severity,
					Message:  fmt.Sprintf("the description of specification %q at location %q does not end with a period", s.Name(), s.Source().Location),
				})
			}
		}
		return result
	}
}
