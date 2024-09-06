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

package specter_test

//func TestPipelineBuilder_WithSourceLoaders(t *testing.T) {
//	loader := &specter.FileSystemSourceLoader{}
//	b := specter.NewPipeline().WithSourceLoaders(loader)
//
//	require.Contains(t, b.SourceLoadingStage.SourceLoaders, loader)
//}
//
//func TestPipelineBuilder_WithUnitLoaders(t *testing.T) {
//	loader := &specterutils.HCLGenericUnitLoader{}
//	p := specter.NewPipeline().WithUnitLoaders(loader)
//	require.Contains(t, p.UnitLoadingStage.Loaders, loader)
//}
//
//func TestPipelineBuilder_WithProcessors(t *testing.T) {
//	processor := specterutils.LintingProcessor{}
//	p := specter.NewPipeline().WithUnitProcessors(processor)
//	require.Contains(t, p.UnitProcessingStage.Processors, processor)
//}
//
//func TestPipelineBuilder_WithArtifactProcessors(t *testing.T) {
//	processor := specter.FileArtifactProcessor{}
//	p := specter.NewPipeline().WithArtifactProcessors(processor)
//	require.Contains(t, p.ArtifactProcessingStage.ArtifactProcessors, processor)
//}
//
//func TestPipelineBuilder_WithArtifactRegistry(t *testing.T) {
//	registry := &specter.InMemoryArtifactRegistry{}
//	p := specter.NewPipeline().WithArtifactRegistry(registry)
//	require.Equal(t, p.ArtifactProcessingStage.ArtifactRegistry, registry)
//}
//
//func TestPipelineBuilder_WithJSONArtifactRegistry(t *testing.T) {
//	fs := &testutils.MockFileSystem{}
//	filePath := specter.DefaultJSONArtifactRegistryFileName
//
//	p := specter.NewPipeline().WithJSONArtifactRegistry(filePath, fs)
//	require.IsType(t, &specter.JSONArtifactRegistry{}, p.ArtifactProcessingStage.ArtifactRegistry)
//	registry := p.ArtifactProcessingStage.ArtifactRegistry.(*specter.JSONArtifactRegistry)
//
//	assert.Equal(t, registry.FileSystem, fs)
//	assert.Equal(t, registry.FilePath, filePath)
//
//	assert.NotNil(t, registry.TimeProvider())
//}
