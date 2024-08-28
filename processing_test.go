package specter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

// MockArtifactRegistry is a mock implementation of ArtifactRegistry
type MockArtifactRegistry struct {
	mock.Mock
}

func (m *MockArtifactRegistry) Load() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockArtifactRegistry) Save() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockArtifactRegistry) Add(processorName string, artifactID ArtifactID) {
	m.Called(processorName, artifactID)
}

func (m *MockArtifactRegistry) Remove(processorName string, artifactID ArtifactID) {
	m.Called(processorName, artifactID)
}

func (m *MockArtifactRegistry) Artifacts(processorName string) []ArtifactID {
	args := m.Called(processorName)
	return args.Get(0).([]ArtifactID)
}

func TestNoopArtifactRegistry_Load(t *testing.T) {
	// Arrange
	registry := NoopArtifactRegistry{}

	// Act
	err := registry.Load()

	// Assert
	assert.NoError(t, err, "Load should not return an error")
}

func TestNoopArtifactRegistry_Save(t *testing.T) {
	// Arrange
	registry := NoopArtifactRegistry{}

	// Act
	err := registry.Save()

	// Assert
	assert.NoError(t, err, "Save should not return an error")
}

func TestNoopArtifactRegistry_Add(t *testing.T) {
	registry := NoopArtifactRegistry{}
	err := registry.Add("processor1", ArtifactRegistryEntry{})
	require.NoError(t, err)
}

func TestNoopArtifactRegistry_Remove(t *testing.T) {
	registry := NoopArtifactRegistry{}
	err := registry.Remove("processor1", "artifactFile.txt")
	require.NoError(t, err)
}

func TestNoopArtifactRegistry_FindAll(t *testing.T) {
	registry := NoopArtifactRegistry{}
	artifacts, err := registry.FindAll("processor1")
	require.NoError(t, err)
	require.Nil(t, artifacts, "FindAll should return nil for NoopArtifactRegistry")
}
