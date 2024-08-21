package specter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// MockOutputRegistry is a mock implementation of OutputRegistry
type MockOutputRegistry struct {
	mock.Mock
}

func (m *MockOutputRegistry) Load() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOutputRegistry) Save() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOutputRegistry) AddOutput(processorName string, outputName string) {
	m.Called(processorName, outputName)
}

func (m *MockOutputRegistry) RemoveOutput(processorName string, outputName string) {
	m.Called(processorName, outputName)
}

func (m *MockOutputRegistry) Outputs(processorName string) []string {
	args := m.Called(processorName)
	return args.Get(0).([]string)
}

func TestOutputProcessingContext__AddToRegistry(t *testing.T) {
	// Arrange
	mockRegistry := new(MockOutputRegistry)
	ctx := &OutputProcessingContext{
		outputRegistry: mockRegistry,
		processorName:  "testProcessor",
	}

	outputName := "outputFile.txt"

	mockRegistry.On("AddOutput", "testProcessor", outputName).Return()

	// Act
	ctx.AddToRegistry(outputName)

	// Assert
	mockRegistry.AssertExpectations(t)
}

func TestOutputProcessingContext__RemoveFromRegistry(t *testing.T) {
	// Arrange
	mockRegistry := new(MockOutputRegistry)
	ctx := &OutputProcessingContext{
		outputRegistry: mockRegistry,
		processorName:  "testProcessor",
	}

	outputName := "outputFile.txt"

	mockRegistry.On("RemoveOutput", "testProcessor", outputName).Return()

	// Act
	ctx.RemoveFromRegistry(outputName)

	// Assert
	mockRegistry.AssertExpectations(t)
}

func TestOutputProcessingContext__RegistryOutputs(t *testing.T) {
	// Arrange
	mockRegistry := new(MockOutputRegistry)
	ctx := &OutputProcessingContext{
		outputRegistry: mockRegistry,
		processorName:  "testProcessor",
	}

	expectedOutputs := []string{"file1.txt", "file2.txt"}

	mockRegistry.On("Outputs", "testProcessor").Return(expectedOutputs)

	// Act
	outputs := ctx.RegistryOutputs()

	// Assert
	assert.Equal(t, expectedOutputs, outputs)
	mockRegistry.AssertExpectations(t)
}

func TestNoopOutputRegistry_Load(t *testing.T) {
	// Arrange
	registry := NoopOutputRegistry{}

	// Act
	err := registry.Load()

	// Assert
	assert.NoError(t, err, "Load should not return an error")
}

func TestNoopOutputRegistry_Save(t *testing.T) {
	// Arrange
	registry := NoopOutputRegistry{}

	// Act
	err := registry.Save()

	// Assert
	assert.NoError(t, err, "Save should not return an error")
}

func TestNoopOutputRegistry_AddOutput(t *testing.T) {
	// Arrange
	registry := NoopOutputRegistry{}

	// Act
	registry.AddOutput("processor1", "outputFile.txt")

	// Assert
	// No state to assert since it's a no-op, just ensure it doesn't panic or error.
}

func TestNoopOutputRegistry_RemoveOutput(t *testing.T) {
	// Arrange
	registry := NoopOutputRegistry{}

	// Act
	registry.RemoveOutput("processor1", "outputFile.txt")

	// Assert
	// No state to assert since it's a no-op, just ensure it doesn't panic or error.
}

func TestNoopOutputRegistry_Outputs(t *testing.T) {
	// Arrange
	registry := NoopOutputRegistry{}

	// Act
	outputs := registry.Outputs("processor1")

	// Assert
	assert.Nil(t, outputs, "Outputs should return nil for NoopOutputRegistry")
}
