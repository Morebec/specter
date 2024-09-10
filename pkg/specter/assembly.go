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
	pipeline *DefaultPipeline

	SourceLoaders      []SourceLoader
	UnitLoaders        []UnitLoader
	UnitPreprocessors  []UnitPreprocessor
	UnitProcessors     []UnitProcessor
	ArtifactProcessors []ArtifactProcessor
	ArtifactRegistry   ArtifactRegistry

	SourceLoadingStageHooks      SourceLoadingStageHooks
	UnitLoadingStageHooks        UnitLoadingStageHooks
	UnitPreprocessingStageHooks  UnitPreprocessingStageHooks
	UnitProcessingStageHooks     UnitProcessingStageHooks
	ArtifactProcessingStageHooks ArtifactProcessingStageHooks
}

// NewPipeline creates a new instance of a *Pipeline using the provided options.
func NewPipeline() PipelineBuilder {
	return PipelineBuilder{}
}

// WithSourceLoaders configures the SourceLoader of a Pipeline instance.
func (b PipelineBuilder) WithSourceLoaders(loaders ...SourceLoader) PipelineBuilder {
	b.SourceLoaders = loaders
	return b
}

// WithUnitLoaders configures the UnitLoader of a Pipeline instance.
func (b PipelineBuilder) WithUnitLoaders(loaders ...UnitLoader) PipelineBuilder {
	b.UnitLoaders = loaders
	return b
}

// WithUnitPreprocessors configures the UnitPreprocessors of a Pipeline instance.
func (b PipelineBuilder) WithUnitPreprocessors(preprocessors ...UnitPreprocessor) PipelineBuilder {
	b.UnitPreprocessors = preprocessors
	return b
}

// WithUnitProcessors configures the UnitProcess of a Pipeline instance.
func (b PipelineBuilder) WithUnitProcessors(processors ...UnitProcessor) PipelineBuilder {
	b.UnitProcessors = processors
	return b
}

// WithArtifactProcessors configures the ArtifactProcessor of a Pipeline instance.
func (b PipelineBuilder) WithArtifactProcessors(processors ...ArtifactProcessor) PipelineBuilder {
	b.ArtifactProcessors = processors
	return b
}

// WithArtifactRegistry configures the ArtifactRegistry of a Pipeline instance.
func (b PipelineBuilder) WithArtifactRegistry(r ArtifactRegistry) PipelineBuilder {
	b.ArtifactRegistry = r
	return b
}

func (b PipelineBuilder) WithSourceLoadingStageHooks(h SourceLoadingStageHooks) PipelineBuilder {
	b.SourceLoadingStageHooks = h
	return b
}

func (b PipelineBuilder) WithUnitLoadingStageHooks(h UnitLoadingStageHooks) PipelineBuilder {
	b.UnitLoadingStageHooks = h
	return b
}

func (b PipelineBuilder) WithUnitPreprocessingStageHooks(hooks UnitPreprocessingStageHooks) PipelineBuilder {
	b.UnitPreprocessingStageHooks = hooks
	return b
}

func (b PipelineBuilder) WithUnitProcessingStageHooks(h UnitProcessingStageHooks) PipelineBuilder {
	b.UnitProcessingStageHooks = h
	return b
}

func (b PipelineBuilder) WithArtifactProcessingStageHooks(h ArtifactProcessingStageHooks) PipelineBuilder {
	b.ArtifactProcessingStageHooks = h
	return b
}

func (b PipelineBuilder) Build() Pipeline {
	return DefaultPipeline{
		TimeProvider: CurrentTimeProvider,
		SourceLoadingStage: sourceLoadingStage{
			SourceLoaders: b.SourceLoaders,
			Hooks:         b.SourceLoadingStageHooks,
		},
		UnitPreprocessingStage: unitPreprocessingStage{
			Preprocessors: b.UnitPreprocessors,
			Hooks:         b.UnitPreprocessingStageHooks,
		},
		UnitLoadingStage: unitLoadingStage{
			Loaders: b.UnitLoaders,
			Hooks:   b.UnitLoadingStageHooks,
		},
		UnitProcessingStage: unitProcessingStage{
			Processors: b.UnitProcessors,
			Hooks:      b.UnitProcessingStageHooks,
		},
		ArtifactProcessingStage: artifactProcessingStage{
			Registry:   b.ArtifactRegistry,
			Processors: b.ArtifactProcessors,
			Hooks:      b.ArtifactProcessingStageHooks,
		},
	}
}

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

// UNIT PROCESSING

func NewUnitProcessorFunc(name string, processFunc func(ctx UnitProcessingContext) ([]Artifact, error)) UnitProcessor {
	return &UnitProcessorFunc{name: name, processFunc: processFunc}
}

// ARTIFACT PROCESSING

func NewArtifactProcessorFunc(name string, processFunc func(ctx ArtifactProcessingContext) error) ArtifactProcessor {
	return &ArtifactProcessorFunc{name: name, processFunc: processFunc}
}

// NewJSONArtifactRegistry returns a new artifact file registry.
func NewJSONArtifactRegistry(fileName string, fs FileSystem) *JSONArtifactRegistry {
	return &JSONArtifactRegistry{
		InMemoryArtifactRegistry: &InMemoryArtifactRegistry{},
		FilePath:                 fileName,
		TimeProvider:             CurrentTimeProvider,
		FileSystem:               fs,
	}
}
