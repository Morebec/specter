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

import (
	"github.com/morebec/specter/pkg/specter"
	"github.com/morebec/specter/pkg/specterutils"
	"github.com/morebec/specter/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWithDefaultLogger(t *testing.T) {
	p := specter.NewPipeline(specter.WithDefaultLogger())
	assert.IsType(t, &specter.DefaultLogger{}, p.Logger)
}

func TestWithSourceLoaders(t *testing.T) {
	loader := &specter.FileSystemSourceLoader{}
	p := specter.NewPipeline(specter.WithSourceLoaders(loader))
	require.Contains(t, p.SourceLoaders, loader)
}

func TestWithLoaders(t *testing.T) {
	loader := &specterutils.HCLGenericUnitLoader{}
	p := specter.NewPipeline(specter.WithLoaders(loader))
	require.Contains(t, p.Loaders, loader)
}

func TestWithProcessors(t *testing.T) {
	processor := specterutils.LintingProcessor{}
	p := specter.NewPipeline(specter.WithProcessors(processor))
	require.Contains(t, p.Processors, processor)
}

func TestWithArtifactProcessors(t *testing.T) {
	processor := specter.FileArtifactProcessor{}
	p := specter.NewPipeline(specter.WithArtifactProcessors(processor))
	require.Contains(t, p.ArtifactProcessors, processor)
}

func TestWithTimeProvider(t *testing.T) {
	tp := specter.CurrentTimeProvider()
	p := specter.NewPipeline(specter.WithTimeProvider(tp))
	require.NotNil(t, p.TimeProvider)
}

func TestWithArtifactRegistry(t *testing.T) {
	registry := &specter.InMemoryArtifactRegistry{}
	p := specter.NewPipeline(specter.WithArtifactRegistry(registry))
	require.Equal(t, p.ArtifactRegistry, registry)
}

func TestWithJSONArtifactRegistry(t *testing.T) {
	fs := &testutils.MockFileSystem{}
	filePath := specter.DefaultJSONArtifactRegistryFileName

	p := specter.NewPipeline(specter.WithJSONArtifactRegistry(filePath, fs))
	require.IsType(t, &specter.JSONArtifactRegistry{}, p.ArtifactRegistry)
	registry := p.ArtifactRegistry.(*specter.JSONArtifactRegistry)

	assert.Equal(t, registry.FileSystem, fs)
	assert.Equal(t, registry.FilePath, filePath)

	assert.NotNil(t, registry.TimeProvider())
}
