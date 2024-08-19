package specter

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"strings"
)

// UndefinedSpecificationName constant used to test against undefined SpecificationName.
const UndefinedSpecificationName SpecificationName = ""

const LintingErrorCode = "linting_error"

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

func (l LintingProcessor) Process(ctx ProcessingContext) ([]ProcessingOutput, error) {
	linter := CompositeSpecificationLinter(l.linters...)
	ctx.Logger.Info("\nLinting specifications ...")
	lr := linter.Lint(ctx.Specifications)
	if lr.HasWarnings() {
		for _, w := range lr.Warnings() {
			ctx.Logger.Warning(fmt.Sprintf("Warning: %s\n", w.Message))
		}
	}

	if lr.HasErrors() {
		for _, e := range lr.Errors().Errors {
			ctx.Logger.Error(fmt.Sprintf("Error: %s\n", e.Error()))
		}
	}

	if !lr.HasWarnings() && !lr.HasErrors() {
		ctx.Logger.Success("Specifications linted successfully.")
	}

	return nil, nil

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
	Lint(specifications SpecificationGroup) LinterResultSet
}

// SpecificationLinterFunc implementation of a SpecificationLinter that relies on a func
type SpecificationLinterFunc func(specifications SpecificationGroup) LinterResultSet

func (l SpecificationLinterFunc) Lint(specifications SpecificationGroup) LinterResultSet {
	return l(specifications)
}

// CompositeSpecificationLinter A Composite linter is responsible for running multiple linters as one.
func CompositeSpecificationLinter(linters ...SpecificationLinter) SpecificationLinterFunc {
	return func(specifications SpecificationGroup) LinterResultSet {
		result := LinterResultSet{}
		for _, linter := range linters {
			lr := linter.Lint(specifications)
			result = append(result, lr...)
		}
		return result
	}
}

// SpecificationMustNotHaveUndefinedNames ensures that no specification has an undefined type FilePath
func SpecificationMustNotHaveUndefinedNames() SpecificationLinterFunc {
	return func(specifications SpecificationGroup) LinterResultSet {
		result := LinterResultSet{}

		for _, s := range specifications {
			if s.Name() == UndefinedSpecificationName {
				result = append(result, LinterResult{
					Severity: ErrorSeverity,
					Message:  fmt.Sprintf("specification at %q has an undefined type FilePath", s.Source().Location),
				})
			}
		}

		return result
	}
}

// SpecificationsMustHaveUniqueNames ensures that names are unique amongst specifications.
func SpecificationsMustHaveUniqueNames() SpecificationLinterFunc {
	return func(specifications SpecificationGroup) LinterResultSet {

		result := LinterResultSet{}

		// Where key is the type FilePath and the array contains all the specification file locations where it was encountered.
		encounteredNames := map[SpecificationName][]string{}

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
				for fn, _ := range fnMap {
					fileNames = append(fileNames, fn)
				}

				result = append(result, LinterResult{
					Severity: ErrorSeverity,
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
func SpecificationsMustHaveDescriptionAttribute() SpecificationLinterFunc {
	return func(specifications SpecificationGroup) LinterResultSet {
		result := LinterResultSet{}
		for _, s := range specifications {
			if s.Description() == "" {
				result = append(result, LinterResult{
					Severity: ErrorSeverity,
					Message:  fmt.Sprintf("specification at location %q does not have a description.", s.Source().Location),
				})
			}
		}
		return result
	}
}

// SpecificationsMustHaveLowerCaseNames ensures that all specification type names are lower case.
func SpecificationsMustHaveLowerCaseNames() SpecificationLinterFunc {
	return func(specifications SpecificationGroup) LinterResultSet {
		result := LinterResultSet{}
		for _, s := range specifications {
			if strings.ToLower(string(s.Name())) != string(s.Name()) {
				result = append(result, LinterResult{
					Severity: ErrorSeverity,
					Message: fmt.Sprintf(
						fmt.Sprintf("specification names must be lowercase got %q at location %q",
							s.Name(),
							s.Source().Location,
						),
					),
				})
			}
		}
		return result
	}
}
