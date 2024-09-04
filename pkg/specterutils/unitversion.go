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
	"github.com/morebec/specter/pkg/specter"
)

type UnitVersion string

type HasVersion interface {
	specter.Unit

	Version() UnitVersion
}

func HasVersionMustHaveAVersionLinter(severity LinterResultSeverity) UnitLinter {
	return UnitLinterFunc(func(units specter.UnitGroup) LinterResultSet {
		var r LinterResultSet
		specs := units.Select(func(u specter.Unit) bool {
			if _, ok := u.(HasVersion); ok {
				return true
			}
			return false
		})

		for _, unit := range specs {
			u := unit.(HasVersion)
			if u.Version() != "" {
				continue
			}

			r = append(r, LinterResult{
				Severity: severity,
				Message:  fmt.Sprintf("unit %q at %q should have a version", unit.Name(), unit.Source().Location),
			})
		}
		return r
	})
}
