package specter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockArtifactRegistry) AddArtifact(processorName string, artifactName string) {
	m.Called(processorName, artifactName)
}

func (m *MockArtifactRegistry) RemoveArtifact(processorName string, artifactName string) {
	m.Called(processorName, artifactName)
}

func (m *MockArtifactRegistry) Artifacts(processorName string) []string {
	args := m.Called(processorName)
	return args.Get(0).([]string)
}

func TestArtifactProcessingContext__AddToRegistry(t *testing.T) {
	// Arrange
	mockRegistry := new(MockArtifactRegistry)
	ctx := &ArtifactProcessingContext{
		artifactRegistry: mockRegistry,
		processorName:    "testProcessor",
	}

	artifactName := "artifactFile.txt"

	mockRegistry.On("AddArtifact", "testProcessor", artifactName).Return()

	// Act
	ctx.AddToRegistry(artifactName)

	// Assert
	mockRegistry.AssertExpectations(t)
}

func TestArtifactProcessingContext__RemoveFromRegistry(t *testing.T) {
	// Arrange
	mockRegistry := new(MockArtifactRegistry)
	ctx := &ArtifactProcessingContext{
		artifactRegistry: mockRegistry,
		processorName:    "testProcessor",
	}

	artifactName := "artifactFile.txt"

	mockRegistry.On("RemoveArtifact", "testProcessor", artifactName).Return()

	// Act
	ctx.RemoveFromRegistry(artifactName)

	// Assert
	mockRegistry.AssertExpectations(t)
}

func TestArtifactProcessingContext__RegistryArtifacts(t *testing.T) {
	// Arrange
	mockRegistry := new(MockArtifactRegistry)
	ctx := &ArtifactProcessingContext{
		artifactRegistry: mockRegistry,
		processorName:    "testProcessor",
	}

	expectedArtifacts := []string{"file1.txt", "file2.txt"}

	mockRegistry.On("Artifacts", "testProcessor").Return(expectedArtifacts)

	// Act
	artifacts := ctx.RegistryArtifacts()

	// Assert
	assert.Equal(t, expectedArtifacts, artifacts)
	mockRegistry.AssertExpectations(t)
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

func TestNoopArtifactRegistry_AddArtifact(t *testing.T) {
	// Arrange
	registry := NoopArtifactRegistry{}

	// Act
	registry.AddArtifact("processor1", "artifactFile.txt")

	// Assert
	// No state to assert since it's a no-op, just ensure it doesn't panic or error.
}

func TestNoopArtifactRegistry_RemoveArtifact(t *testing.T) {
	// Arrange
	registry := NoopArtifactRegistry{}

	// Act
	registry.RemoveArtifact("processor1", "artifactFile.txt")

	// Assert
	// No state to assert since it's a no-op, just ensure it doesn't panic or error.
}

func TestNoopArtifactRegistry_Artifacts(t *testing.T) {
	// Arrange
	registry := NoopArtifactRegistry{}

	// Act
	artifacts := registry.Artifacts("processor1")

	// Assert
	assert.Nil(t, artifacts, "Artifacts should return nil for NoopArtifactRegistry")
}
