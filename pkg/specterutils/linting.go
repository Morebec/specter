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
	"io"
	"strings"
	"unicode"
)

const LinterResultArtifactID = "_linting_processor_results"

// UndefinedUnitName constant used to test against undefined UnitName.
const UndefinedUnitName specter.UnitName = ""

const LintingErrorCode = "specter.spec_processing.linting_error"

type LinterResultSeverity string

const (
	ErrorSeverity   LinterResultSeverity = "error"
	WarningSeverity LinterResultSeverity = "warning"
)

type LintingProcessor struct {
	linters []UnitLinter
	Logger  specter.Logger
}

func NewLintingProcessor(linters ...UnitLinter) *LintingProcessor {
	return &LintingProcessor{linters: linters}
}

func (l LintingProcessor) Name() string {
	return "linting_processor"
}

func (l LintingProcessor) Process(ctx specter.ProcessingContext) (artifacts []specter.Artifact, err error) {
	if l.Logger == nil {
		l.Logger = specter.NewDefaultLogger(specter.DefaultLoggerConfig{Writer: io.Discard})
	}

	linter := CompositeUnitLinter(l.linters...)

	lr := linter.Lint(ctx.Units)

	artifacts = append(artifacts, lr)

	if lr.HasWarnings() {
		for _, w := range lr.Warnings() {
			l.Logger.Warning(fmt.Sprintf("Warning: %s\n", w.Message))
		}
	}

	if lr.HasErrors() {
		for _, e := range lr.Errors().Errors {
			l.Logger.Error(fmt.Sprintf("Error: %s\n", e.Error()))
		}
		err = lr.Errors()
	}

	if !lr.HasWarnings() && !lr.HasErrors() {
		l.Logger.Success("Units linted successfully.")
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

// UnitLinter represents a function responsible for linting units.
type UnitLinter interface {
	Lint(units specter.UnitGroup) LinterResultSet
}

// UnitLinterFunc implementation of a UnitLinter that relies on a func
type UnitLinterFunc func(units specter.UnitGroup) LinterResultSet

func (l UnitLinterFunc) Lint(units specter.UnitGroup) LinterResultSet {
	return l(units)
}

// CompositeUnitLinter A Composite linter is responsible for running multiple linters as one.
func CompositeUnitLinter(linters ...UnitLinter) UnitLinterFunc {
	return func(units specter.UnitGroup) LinterResultSet {
		var result LinterResultSet
		for _, linter := range linters {
			lr := linter.Lint(units)
			result = append(result, lr...)
		}
		return result
	}
}

// UnitMustNotHaveUndefinedNames ensures that no unit has an undefined name
func UnitMustNotHaveUndefinedNames(severity LinterResultSeverity) UnitLinterFunc {
	return func(units specter.UnitGroup) LinterResultSet {
		var result LinterResultSet

		for _, u := range units {
			if u.Name() == UndefinedUnitName {
				result = append(result, LinterResult{
					Severity: severity,
					Message:  fmt.Sprintf("unit at %q has an undefined name", u.Source().Location),
				})
			}
		}

		return result
	}
}

// UnitsMustHaveUniqueNames ensures that names are unique amongst units.
func UnitsMustHaveUniqueNames(severity LinterResultSeverity) UnitLinterFunc {
	return func(units specter.UnitGroup) LinterResultSet {

		var result LinterResultSet

		// Where key is the type FilePath and the array contains all the unit file locations where it was encountered.
		encounteredNames := map[specter.UnitName][]string{}

		for _, u := range units {
			if _, found := encounteredNames[u.Name()]; found {
				encounteredNames[u.Name()] = append(encounteredNames[u.Name()], u.Source().Location)
			} else {
				encounteredNames[u.Name()] = []string{u.Source().Location}
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
						"duplicate unit name detected for %q in the following file(s): %s",
						name,
						strings.Join(fileNames, ", "),
					),
				})
			}
		}

		return result
	}
}

// UnitsMustHaveDescriptionAttribute ensures that all units have a description.
func UnitsMustHaveDescriptionAttribute(severity LinterResultSeverity) UnitLinterFunc {
	return func(units specter.UnitGroup) LinterResultSet {
		var result LinterResultSet
		for _, u := range units {
			if u.Description() == "" {
				result = append(result, LinterResult{
					Severity: severity,
					Message:  fmt.Sprintf("unit %q at location %q does not have a description", u.Name(), u.Source().Location),
				})
			}
		}
		return result
	}
}

// UnitsDescriptionsMustStartWithACapitalLetter ensures that unit descriptions start with a capital letter.
func UnitsDescriptionsMustStartWithACapitalLetter(severity LinterResultSeverity) UnitLinterFunc {
	return func(units specter.UnitGroup) LinterResultSet {
		var result LinterResultSet
		for _, u := range units {
			if u.Description() != "" {
				firstLetter := rune(u.Description()[0])
				if unicode.IsUpper(firstLetter) {
					continue
				}
			}
			result = append(result, LinterResult{
				Severity: severity,
				Message:  fmt.Sprintf("the description of unit %q at location %q does not start with a capital letter", u.Name(), u.Source().Location),
			})
		}
		return result
	}
}

// UnitsDescriptionsMustEndWithPeriod ensures that unit descriptions end with a period.
func UnitsDescriptionsMustEndWithPeriod(severity LinterResultSeverity) UnitLinterFunc {
	return func(units specter.UnitGroup) LinterResultSet {
		var result LinterResultSet
		for _, u := range units {
			if !strings.HasSuffix(u.Description(), ".") {
				result = append(result, LinterResult{
					Severity: severity,
					Message:  fmt.Sprintf("the description of unit %q at location %q does not end with a period", u.Name(), u.Source().Location),
				})
			}
		}
		return result
	}
}
