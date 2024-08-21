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
