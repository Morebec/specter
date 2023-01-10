package specter

// UnsupportedSpecLoaderCode ErrorSeverity code returned by a SpecLoader when a given loader does not support a certain source.
const UnsupportedSpecLoaderCode = "unsupported_spec_loader"

// SpecLoader is a service responsible for loading Specs from Sources.
type SpecLoader interface {
	// Load loads a slice of Spec from a Source, or returns an error if it encountered a failure.
	Load(s Source) ([]Spec, error)

	// SupportsSource indicates if this loader supports a certain source or not.
	SupportsSource(s Source) bool
}
