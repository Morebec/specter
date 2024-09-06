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
)

const LinterResultArtifactID = "_linting_processor_results"

// UndefinedUnitID constant used to test against undefined specter.UnitID.
const UndefinedUnitID specter.UnitID = ""

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

func (l LintingProcessor) Process(ctx specter.UnitProcessingContext) (artifacts []specter.Artifact, err error) {
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

func GetLintingResultsFromContext(ctx specter.UnitProcessingContext) LinterResultSet {
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

// UnitsMustHaveIDs ensures that no unit has an undefined ID.
func UnitsMustHaveIDs(severity LinterResultSeverity) UnitLinterFunc {
	return func(units specter.UnitGroup) LinterResultSet {
		var result LinterResultSet

		for _, u := range units {
			if u.ID() == UndefinedUnitID {
				result = append(result, LinterResult{
					Severity: severity,
					Message:  fmt.Sprintf("a unit of kind %q has no ID at %q", u.Kind(), u.Source().Location),
				})
			}
		}

		return result
	}
}

// UnitsIDsMustBeUnique ensures that units all have unique IDs.
func UnitsIDsMustBeUnique(severity LinterResultSeverity) UnitLinterFunc {
	return func(units specter.UnitGroup) LinterResultSet {

		var result LinterResultSet

		// Where key is the type ID and the array contains all the unit file locations where it was encountered.
		// TODO simplify
		encounteredIDs := map[specter.UnitID][]string{}

		for _, u := range units {
			if _, found := encounteredIDs[u.ID()]; found {
				encounteredIDs[u.ID()] = append(encounteredIDs[u.ID()], u.Source().Location)
			} else {
				encounteredIDs[u.ID()] = []string{u.Source().Location}
			}
		}

		for id, files := range encounteredIDs {
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
						"duplicate unit ID detected %q in the following file(s): %s",
						id,
						strings.Join(fileNames, ", "),
					),
				})
			}
		}

		return result
	}
}
