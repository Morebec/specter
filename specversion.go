package specter

import "fmt"

type SpecificationVersion string

type HasVersion interface {
	Specification

	Version() SpecificationVersion
}

func HasVersionMustHaveAVersionLinter(severity LinterResultSeverity) SpecificationLinter {
	return SpecificationLinterFunc(func(specifications SpecificationGroup) LinterResultSet {
		var r LinterResultSet
		specs := specifications.Select(func(s Specification) bool {
			if _, ok := s.(HasVersion); ok {
				return true
			}

			return false
		})

		for _, spec := range specs {
			s := spec.(HasVersion)
			if s.Version() != "" {
				continue
			}

			r = append(r, LinterResult{
				Severity: severity,
				Message:  fmt.Sprintf("specification %q at %q should have a version", spec.Name(), spec.Source().Location),
			})
		}
		return r
	})
}
