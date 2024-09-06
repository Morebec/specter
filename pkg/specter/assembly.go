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

type PipelineBuilder struct {
	*DefaultPipeline
}

// NewPipeline creates a new instance of a *Pipeline using the provided options.
func NewPipeline(opts ...PipelineOption) PipelineBuilder {
	return PipelineBuilder{
		DefaultPipeline: &DefaultPipeline{
			TimeProvider:            CurrentTimeProvider,
			sourceLoadingStage:      sourceLoadingStage{},
			unitLoadingStage:        unitLoadingStage{},
			unitProcessingStage:     unitProcessingStage{},
			artifactProcessingStage: artifactProcessingStage{},
		},
	}
}

// PipelineOption represents an option to configure a Pipeline instance.
type PipelineOption func(s *DefaultPipeline)

// WithSourceLoaders configures the SourceLoader of a Pipeline instance.
func (b PipelineBuilder) WithSourceLoaders(loaders ...SourceLoader) PipelineBuilder {
	b.sourceLoadingStage.SourceLoaders = loaders
	return b
}

// WithUnitLoaders configures the UnitLoader of a Pipeline instance.
func (b PipelineBuilder) WithUnitLoaders(loaders ...UnitLoader) PipelineBuilder {
	b.unitLoadingStage.Loaders = loaders
	return b
}

// WithProcessors configures the UnitProcess of a Pipeline instance.
func (b PipelineBuilder) WithProcessors(processors ...UnitProcessor) PipelineBuilder {
	b.unitProcessingStage.Processors = processors
	return b
}

// WithArtifactProcessors configures the ArtifactProcessor of a Pipeline instance.
func (b PipelineBuilder) WithArtifactProcessors(processors ...ArtifactProcessor) PipelineBuilder {
	b.artifactProcessingStage.ArtifactProcessors = processors
	return b
}

// WithArtifactRegistry configures the ArtifactRegistry of a Pipeline instance.
func (b PipelineBuilder) WithArtifactRegistry(r ArtifactRegistry) PipelineBuilder {
	b.artifactProcessingStage.ArtifactRegistry = r
	return b
}

// DEFAULTS PIPELINE OPTIONS

func (b PipelineBuilder) WithJSONArtifactRegistry(fileName string, fs FileSystem) PipelineBuilder {
	return b.WithArtifactRegistry(NewJSONArtifactRegistry(fileName, fs))
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
		TimeProvider:             CurrentTimeProvider,
		FileSystem:               fs,
	}
}
