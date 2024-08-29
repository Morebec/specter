package specter

import (
	"os"
	"time"
)

// New allows creating a new specter instance using the provided options.
func New(opts ...Option) *Specter {
	s := &Specter{
		Logger:        NewDefaultLogger(DefaultLoggerConfig{DisableColors: true, Writer: os.Stdout}),
		ExecutionMode: FullMode,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Option represents an option to configure a specter instance.
type Option func(s *Specter)

// WithLogger configures the Logger of a Specter instance.
func WithLogger(l Logger) Option {
	return func(s *Specter) {
		s.Logger = l
	}
}

// WithSourceLoaders configures the SourceLoader of a Specter instance.
func WithSourceLoaders(loaders ...SourceLoader) Option {
	return func(s *Specter) {
		s.SourceLoaders = append(s.SourceLoaders, loaders...)
	}
}

// WithLoaders configures the SpecificationLoader of a Specter instance.
func WithLoaders(loaders ...SpecificationLoader) Option {
	return func(s *Specter) {
		s.Loaders = append(s.Loaders, loaders...)
	}
}

// WithProcessors configures the SpecProcess of a Specter instance.
func WithProcessors(processors ...SpecificationProcessor) Option {
	return func(s *Specter) {
		s.Processors = append(s.Processors, processors...)
	}
}

// WithArtifactProcessors configures the ArtifactProcessor of a Specter instance.
func WithArtifactProcessors(processors ...ArtifactProcessor) Option {
	return func(s *Specter) {
		s.ArtifactProcessors = append(s.ArtifactProcessors, processors...)
	}
}

// WithExecutionMode configures the ExecutionMode of a Specter instance.
func WithExecutionMode(m ExecutionMode) Option {
	return func(s *Specter) {
		s.ExecutionMode = m
	}
}
func WithArtifactRegistry(r ArtifactRegistry) Option {
	return func(s *Specter) {
		s.ArtifactRegistry = r
	}
}

// DEFAULTS SPECTER OPTIONS

func WithDefaultLogger() Option {
	return WithLogger(NewDefaultLogger(DefaultLoggerConfig{DisableColors: false, Writer: os.Stdout}))
}

func WithJSONArtifactRegistry(fileName string, fs FileSystem) Option {
	return WithArtifactRegistry(NewJSONArtifactRegistry(fileName, fs))
}

// NewFileSystemLoader FileSystem
func NewFileSystemLoader(fs FileSystem) *FileSystemLoader {
	return &FileSystemLoader{fs: fs}
}

func NewLocalFileSourceLoader() FileSystemLoader {
	return FileSystemLoader{fs: LocalFileSystem{}}
}

// ARTIFACT REGISTRIES

// NewJSONArtifactRegistry returns a new artifact file registry.
func NewJSONArtifactRegistry(fileName string, fs FileSystem) *JSONArtifactRegistry {
	return &JSONArtifactRegistry{
		Entries:  make(map[string][]JsonArtifactRegistryEntry),
		FilePath: fileName,
		CurrentTimeProvider: func() time.Time {
			return time.Now()
		},
		FileSystem: fs,
	}
}
