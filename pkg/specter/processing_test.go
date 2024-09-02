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

func TestProcessorArtifactRegistry_Add(t *testing.T) {
	r := &InMemoryArtifactRegistry{}
	pr := ProcessorArtifactRegistry{
		processorName: "unit_tester",
		registry:      r,
	}

	err := pr.Add("an_artifact", nil)
	require.NoError(t, err)

	_, found, err := r.FindByID("unit_tester", "an_artifact")
	require.NoError(t, err)
	require.True(t, found)
}

func TestProcessorArtifactRegistry_Remove(t *testing.T) {
	r := &InMemoryArtifactRegistry{}
	pr := ProcessorArtifactRegistry{
		processorName: "unit_tester",
		registry:      r,
	}

	err := r.Add("unit_tester", ArtifactRegistryEntry{
		ArtifactID: "an_artifact",
		Metadata:   nil,
	})
	require.NoError(t, err)

	err = pr.Remove("an_artifact")
	require.NoError(t, err)

	_, found, err := r.FindByID("unit_tester", "an_artifact")
	require.NoError(t, err)
	require.False(t, found)
}

func TestProcessorArtifactRegistry_FindByID(t *testing.T) {
	r := &InMemoryArtifactRegistry{}
	pr := ProcessorArtifactRegistry{
		processorName: "unit_tester",
		registry:      r,
	}

	err := r.Add("unit_tester", ArtifactRegistryEntry{
		ArtifactID: "an_artifact",
		Metadata:   nil,
	})
	require.NoError(t, err)

	_, found, err := pr.FindByID("an_artifact")
	require.NoError(t, err)
	require.True(t, found)
}
