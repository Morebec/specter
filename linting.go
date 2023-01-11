package specter

import (
	"fmt"
	"github.com/morebec/errors-go/errors"
	"strings"
)

// UndefinedSpecificationName constant used to test against undefined SpecName.
const UndefinedSpecificationName SpecName = ""

const LintingErrorCode = "linting_error"

type LinterResultSeverity string

const (
	ErrorSeverity   LinterResultSeverity = "error"
	WarningSeverity LinterResultSeverity = "warning"
)

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

// SpecLinter represents a function responsible for linting specs.
type SpecLinter interface {
	Lint(specs SpecGroup) LinterResultSet
}

// LinterFunc implementation of a SpecLinter that relies on a func
type LinterFunc func(specs SpecGroup) LinterResultSet

func (l LinterFunc) Lint(specs SpecGroup) LinterResultSet {
	return l(specs)
}

// CompositeLinter A Composite linter is responsible for running multiple linters as one.
func CompositeLinter(linters ...SpecLinter) LinterFunc {
	return func(specs SpecGroup) LinterResultSet {
		result := LinterResultSet{}
		for _, linter := range linters {
			lr := linter.Lint(specs)
			result = append(result, lr...)
		}
		return result
	}
}

// SpecificationMustNotHaveUndefinedNames ensures that no spec has an undefined type FilePath
func SpecificationMustNotHaveUndefinedNames() LinterFunc {
	return func(specs SpecGroup) LinterResultSet {
		result := LinterResultSet{}

		for _, s := range specs {
			if s.Name() == UndefinedSpecificationName {
				result = append(result, LinterResult{
					Severity: ErrorSeverity,
					Message:  fmt.Sprintf("spec at \"%s\" has an undefined type FilePath", s.Source().Location),
				})
			}
		}

		return result
	}
}

// SpecificationsMustNotHaveDuplicateTypeNames ensures that type names are unique amongst specs.
func SpecificationsMustNotHaveDuplicateTypeNames() LinterFunc {
	return func(specs SpecGroup) LinterResultSet {

		result := LinterResultSet{}

		// Where key is the type FilePath and the array contains all the spec file locations where it was encountered.
		encounteredNames := map[SpecName][]string{}

		for _, s := range specs {
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
						"duplicate spec FilePath detected for \"%s\" in the following file(s): %s",
						name,
						strings.Join(fileNames, ", "),
					),
				})
			}
		}

		return result
	}
}

// SpecificationsMustHaveDescriptionAttribute ensures that all specs have a description.
func SpecificationsMustHaveDescriptionAttribute() LinterFunc {
	return func(specs SpecGroup) LinterResultSet {
		result := LinterResultSet{}
		for _, s := range specs {
			if s.Description() == "" {
				result = append(result, LinterResult{
					Severity: ErrorSeverity,
					Message:  fmt.Sprintf("spec at location \"%s\" does not have a description.", s.Source().Location),
				})
			}
		}
		return result
	}
}

// SpecificationsMustHaveLowerCaseNames ensures that all spec type names are lower case.
func SpecificationsMustHaveLowerCaseNames() LinterFunc {
	return func(specs SpecGroup) LinterResultSet {
		result := LinterResultSet{}
		for _, s := range specs {
			if strings.ToLower(string(s.Name())) != string(s.Name()) {
				result = append(result, LinterResult{
					Severity: ErrorSeverity,
					Message: fmt.Sprintf(
						fmt.Sprintf("spec type names must be lowercase got \"%s\" at location \"%s\"",
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
