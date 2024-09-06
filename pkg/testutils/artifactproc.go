package testutils

import (
	"github.com/morebec/specter/pkg/specter"
	"github.com/stretchr/testify/mock"
)

type ArtifactStub struct {
	id specter.ArtifactID
}

func NewArtifactStub(id specter.ArtifactID) *ArtifactStub {
	return &ArtifactStub{id: id}
}

func (m ArtifactStub) ID() specter.ArtifactID {
	return m.id
}

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

func (m *MockArtifactRegistry) Add(processorName string, artifactID specter.ArtifactID) {
	m.Called(processorName, artifactID)
}

func (m *MockArtifactRegistry) Remove(processorName string, artifactID specter.ArtifactID) {
	m.Called(processorName, artifactID)
}

func (m *MockArtifactRegistry) Artifacts(processorName string) []specter.ArtifactID {
	args := m.Called(processorName)
	return args.Get(0).([]specter.ArtifactID)
}
