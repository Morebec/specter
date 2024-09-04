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
	. "github.com/morebec/specter/pkg/specter"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProcessorArtifactRegistry_Add(t *testing.T) {
	r := &InMemoryArtifactRegistry{}
	pr := NewProcessorArtifactRegistry("unit_tester", r)

	err := pr.Add("an_artifact", nil)
	require.NoError(t, err)

	_, found, err := r.FindByID("unit_tester", "an_artifact")
	require.NoError(t, err)
	require.True(t, found)
}

func TestProcessorArtifactRegistry_Remove(t *testing.T) {
	r := &InMemoryArtifactRegistry{}
	pr := NewProcessorArtifactRegistry("unit_tester", r)

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
	pr := NewProcessorArtifactRegistry("unit_tester", r)

	err := r.Add("unit_tester", ArtifactRegistryEntry{
		ArtifactID: "an_artifact",
		Metadata:   nil,
	})
	require.NoError(t, err)

	_, found, err := pr.FindByID("an_artifact")
	require.NoError(t, err)
	require.True(t, found)
}

func TestGetContextArtifact(t *testing.T) {
	type when struct {
		ctx ProcessingContext
		id  ArtifactID
	}
	type then[T Artifact] struct {
		artifact T
	}
	type testCase[T Artifact] struct {
		name string
		when when
		then then[T]
	}
	tests := []testCase[*FileArtifact]{
		{
			name: "GIVEN no artifact matches THEN return nil",
			when: when{
				ctx: ProcessingContext{},
				id:  "not_found",
			},
			then: then[*FileArtifact]{
				artifact: nil,
			},
		},
		{
			name: "GIVEN artifact matches THEN return artifact",
			when: when{
				ctx: ProcessingContext{
					Artifacts: []Artifact{
						&FileArtifact{Path: "/path/to/file"},
					},
				},
				id: "/path/to/file",
			},
			then: then[*FileArtifact]{
				artifact: &FileArtifact{Path: "/path/to/file"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualArtifact := GetContextArtifact[*FileArtifact](tt.when.ctx, tt.when.id)
			require.Equal(t, tt.then.artifact, actualArtifact)
		})
	}
}
