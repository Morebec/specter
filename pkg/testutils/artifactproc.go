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
