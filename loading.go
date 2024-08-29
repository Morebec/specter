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

// UnsupportedSourceErrorCode ErrorSeverity code returned by a SpecificationLoader when a given loader does not support a certain source.
const UnsupportedSourceErrorCode = "unsupported_source"

// SpecificationLoader is a service responsible for loading Specifications from Sources.
type SpecificationLoader interface {
	// Load loads a slice of Specification from a Source, or returns an error if it encountered a failure.
	Load(s Source) ([]Specification, error)

	// SupportsSource indicates if this loader supports a certain source or not.
	SupportsSource(s Source) bool
}
