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
