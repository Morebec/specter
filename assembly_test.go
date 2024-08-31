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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWithDefaultLogger(t *testing.T) {
	s := New(WithDefaultLogger())
	assert.IsType(t, &DefaultLogger{}, s.Logger)
}

func TestWithSourceLoaders(t *testing.T) {
	loader := &FileSystemSourceLoader{}
	s := New(WithSourceLoaders(loader))
	require.Contains(t, s.SourceLoaders, loader)
}

func TestWithLoaders(t *testing.T) {
	loader := &HCLGenericSpecLoader{}
	s := New(WithLoaders(loader))
	require.Contains(t, s.Loaders, loader)
}

func TestWithProcessors(t *testing.T) {
	processor := LintingProcessor{}
	s := New(WithProcessors(processor))
	require.Contains(t, s.Processors, processor)
}

func TestWithArtifactProcessors(t *testing.T) {
	processor := FileArtifactProcessor{}
	s := New(WithArtifactProcessors(processor))
	require.Contains(t, s.ArtifactProcessors, processor)
}

func TestWithExecutionMode(t *testing.T) {
	mode := PreviewMode
	s := New(WithExecutionMode(mode))
	require.Equal(t, s.ExecutionMode, mode)
}

func TestWithArtifactRegistry(t *testing.T) {
	registry := &InMemoryArtifactRegistry{}
	s := New(WithArtifactRegistry(registry))
	require.Equal(t, s.ArtifactRegistry, registry)
}

func TestWithJSONArtifactRegistry(t *testing.T) {
	fs := &mockFileSystem{}
	filePath := DefaultJSONArtifactRegistryFileName

	s := New(WithJSONArtifactRegistry(filePath, fs))
	require.IsType(t, &JSONArtifactRegistry{}, s.ArtifactRegistry)
	registry := s.ArtifactRegistry.(*JSONArtifactRegistry)

	assert.Equal(t, registry.FileSystem, fs)
	assert.Equal(t, registry.FilePath, filePath)

	assert.NotNil(t, registry.TimeProvider())
}
