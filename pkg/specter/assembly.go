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

import (
	"os"
)

// NewPipeline creates a new instance of a *Pipeline using the provided options.
func NewPipeline(opts ...PipelineOption) *Pipeline {
	s := &Pipeline{
		Logger:       NewDefaultLogger(DefaultLoggerConfig{DisableColors: false, Writer: os.Stdout}),
		TimeProvider: CurrentTimeProvider(),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// PipelineOption represents an option to configure a Pipeline instance.
type PipelineOption func(s *Pipeline)

// WithLogger configures the Logger of a Pipeline instance.
func WithLogger(l Logger) PipelineOption {
	return func(p *Pipeline) {
		p.Logger = l
	}
}

// WithSourceLoaders configures the SourceLoader of a Pipeline instance.
func WithSourceLoaders(loaders ...SourceLoader) PipelineOption {
	return func(p *Pipeline) {
		p.SourceLoaders = append(p.SourceLoaders, loaders...)
	}
}

// WithLoaders configures the UnitLoader of a Pipeline instance.
func WithLoaders(loaders ...UnitLoader) PipelineOption {
	return func(p *Pipeline) {
		p.Loaders = append(p.Loaders, loaders...)
	}
}

// WithProcessors configures the UnitProcess of a Pipeline instance.
func WithProcessors(processors ...UnitProcessor) PipelineOption {
	return func(p *Pipeline) {
		p.Processors = append(p.Processors, processors...)
	}
}

// WithArtifactProcessors configures the ArtifactProcessor of a Pipeline instance.
func WithArtifactProcessors(processors ...ArtifactProcessor) PipelineOption {
	return func(p *Pipeline) {
		p.ArtifactProcessors = append(p.ArtifactProcessors, processors...)
	}
}

// WithTimeProvider configures the TimeProvider of a Pipeline instance.
func WithTimeProvider(tp TimeProvider) PipelineOption {
	return func(p *Pipeline) {
		p.TimeProvider = tp
	}
}

// WithArtifactRegistry configures the ArtifactRegistry of a Pipeline instance.
func WithArtifactRegistry(r ArtifactRegistry) PipelineOption {
	return func(p *Pipeline) {
		p.ArtifactRegistry = r
	}
}

// DEFAULTS PIPELINE OPTIONS

func WithDefaultLogger() PipelineOption {
	return WithLogger(NewDefaultLogger(DefaultLoggerConfig{DisableColors: false, Writer: os.Stdout}))
}

func WithJSONArtifactRegistry(fileName string, fs FileSystem) PipelineOption {
	return WithArtifactRegistry(NewJSONArtifactRegistry(fileName, fs))
}

// Loaders

// NewFileSystemSourceLoader constructs a FileSystemSourceLoader that uses a given FileSystem.
func NewFileSystemSourceLoader(fs FileSystem) *FileSystemSourceLoader {
	return &FileSystemSourceLoader{fs: fs}
}

// NewLocalFileSourceLoader returns a new FileSystemSourceLoader that uses a LocalFileSystem.
func NewLocalFileSourceLoader() *FileSystemSourceLoader {
	return NewFileSystemSourceLoader(LocalFileSystem{})
}

// ARTIFACT REGISTRIES

// NewJSONArtifactRegistry returns a new artifact file registry.
func NewJSONArtifactRegistry(fileName string, fs FileSystem) *JSONArtifactRegistry {
	return &JSONArtifactRegistry{
		InMemoryArtifactRegistry: &InMemoryArtifactRegistry{},
		FilePath:                 fileName,
		TimeProvider:             CurrentTimeProvider(),
		FileSystem:               fs,
	}
}
