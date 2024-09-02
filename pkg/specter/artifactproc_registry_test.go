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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestImplementations(t *testing.T) {
	// InMemoryArtifactRegistry
	assertArtifactRegistryCompliance(t, "InMemoryArtifactRegistry", func() *InMemoryArtifactRegistry {
		return &InMemoryArtifactRegistry{}
	})

	// JSON Artifact Registry
	assertArtifactRegistryCompliance(t, "JSONArtifactRegistry", func() *JSONArtifactRegistry {
		return NewJSONArtifactRegistry(DefaultJSONArtifactRegistryFileName, &mockFileSystem{})
	})
}

func assertArtifactRegistryCompliance[T ArtifactRegistry](t *testing.T, name string, new func() T) {

	it := func(methodName, testName string, testFunc func(t *testing.T)) {
		t.Run(name+"__"+methodName+"_"+testName, testFunc)
	}

	// Add
	it("Add", "should allow adding an entry", func(t *testing.T) {
		r := new()
		err := r.Add("unit_tester", ArtifactRegistryEntry{
			ArtifactID: "an_artifact",
		})
		require.NoError(t, err)

		all, err := r.FindAll("unit_tester")
		require.NoError(t, err)
		assert.Equal(t, []ArtifactRegistryEntry{
			{
				ArtifactID: "an_artifact",
			},
		}, all)
	})

	it("Add", "should not allow adding an entry without an ID", func(t *testing.T) {
		r := new()
		err := r.Add("unit_tester", ArtifactRegistryEntry{
			ArtifactID: "",
		})
		require.Error(t, err)
	})

	it("Add", "should not allow adding an entry without a processor name", func(t *testing.T) {
		r := new()
		err := r.Add("", ArtifactRegistryEntry{
			ArtifactID: "an_artifact",
		})
		require.Error(t, err)
	})

	it("Add", "should allow adding multiple entries with the same processor", func(t *testing.T) {
		r := new()
		err := r.Add("unit_tester", ArtifactRegistryEntry{
			ArtifactID: "artifact_one",
		})
		require.NoError(t, err)

		err = r.Add("unit_tester", ArtifactRegistryEntry{
			ArtifactID: "artifact_two",
		})
		require.NoError(t, err)

		all, err := r.FindAll("unit_tester")
		require.NoError(t, err)
		assert.ElementsMatch(t, []ArtifactRegistryEntry{
			{ArtifactID: "artifact_one"},
			{ArtifactID: "artifact_two"},
		}, all)
	})

	it("Add", "should allow adding entries with different processors", func(t *testing.T) {
		r := new()
		err := r.Add("processor_one", ArtifactRegistryEntry{
			ArtifactID: "artifact_one",
		})
		require.NoError(t, err)

		err = r.Add("processor_two", ArtifactRegistryEntry{
			ArtifactID: "artifact_two",
		})
		require.NoError(t, err)

		allOne, err := r.FindAll("processor_one")
		require.NoError(t, err)
		assert.Equal(t, []ArtifactRegistryEntry{
			{ArtifactID: "artifact_one"},
		}, allOne)

		allTwo, err := r.FindAll("processor_two")
		require.NoError(t, err)
		assert.Equal(t, []ArtifactRegistryEntry{
			{ArtifactID: "artifact_two"},
		}, allTwo)
	})

	it("Add", "should allow adding duplicate entries with idempotency", func(t *testing.T) {
		r := new()
		err := r.Add("unit_tester", ArtifactRegistryEntry{
			ArtifactID: "duplicate_artifact",
		})
		require.NoError(t, err)

		err = r.Add("unit_tester", ArtifactRegistryEntry{
			ArtifactID: "duplicate_artifact",
		})
		require.NoError(t, err)

		allTwo, err := r.FindAll("unit_tester")
		require.NoError(t, err)
		assert.Equal(t, []ArtifactRegistryEntry{
			{ArtifactID: "duplicate_artifact"},
		}, allTwo)
	})

	it("Add", "should allow adding an entry with special characters in the ID", func(t *testing.T) {
		r := new()
		specialID := "artifact-!@#$%^&*()_+=-{}[]|\\:;\"'<>,.?/"
		err := r.Add("unit-tester", ArtifactRegistryEntry{
			ArtifactID: ArtifactID(specialID),
		})
		require.NoError(t, err)

		all, err := r.FindAll("unit-tester")
		require.NoError(t, err)
		assert.Equal(t, []ArtifactRegistryEntry{
			{ArtifactID: ArtifactID(specialID)},
		}, all)
	})

	// Remove
	it("Remove", "should allow removing an entry", func(t *testing.T) {
		r := new()
		err := r.Add("unit_tester", ArtifactRegistryEntry{
			ArtifactID: "an_artifact",
		})
		require.NoError(t, err)

		err = r.Remove("unit_tester", "an_artifact")
		require.NoError(t, err)

		all, err := r.FindAll("unit_tester")
		require.NoError(t, err)
		assert.Nil(t, all)
	})

	it("Remove", "should allow removing an entry that does not exist without having an error", func(t *testing.T) {
		r := new()
		err := r.Remove("unit_tester", "nonexistent_artifact")
		require.NoError(t, err)
	})

	it("Remove", "should not allow removing an entry when the processor name is different", func(t *testing.T) {
		r := new()
		err := r.Add("unit_tester", ArtifactRegistryEntry{
			ArtifactID: "artifact_to_remove",
		})
		require.NoError(t, err)

		err = r.Remove("different)tester", "artifact_to_remove")
		require.NoError(t, err)

		all, err := r.FindAll("unit_tester")
		require.NoError(t, err)
		assert.Equal(t, []ArtifactRegistryEntry{
			{ArtifactID: "artifact_to_remove"},
		}, all)
	})

	it("Remove", "should allow removing multiple times without side effects", func(t *testing.T) {
		r := new()
		err := r.Add("unit_tester", ArtifactRegistryEntry{
			ArtifactID: "an_artifact",
		})
		require.NoError(t, err)

		err = r.Remove("unit_tester", "an_artifact")
		require.NoError(t, err)

		all, err := r.FindAll("unit_tester")
		require.NoError(t, err)
		assert.Nil(t, all)

		err = r.Remove("unit_tester", "an_artifact")
		require.NoError(t, err)

		all, err = r.FindAll("unit_tester")
		require.NoError(t, err)
		assert.Nil(t, all)
	})

	it("Remove", "should not allow removing an entry without an ID", func(t *testing.T) {
		r := new()

		err := r.Remove("unit_tester", "")
		require.Error(t, err)
	})

	it("Remove", "should not allow removing an entry without a processor name", func(t *testing.T) {
		r := new()

		err := r.Remove("", "an_artifact")
		require.Error(t, err)
	})

	it("Remove", "should return an empty list after removing the last entry", func(t *testing.T) {
		r := new()
		err := r.Add("unit-tester", ArtifactRegistryEntry{
			ArtifactID: "last_artifact",
		})
		require.NoError(t, err)

		err = r.Remove("unit-tester", "last_artifact")
		require.NoError(t, err)

		all, err := r.FindAll("unit-tester")
		require.NoError(t, err)
		assert.Nil(t, all)
	})

	// Find all
	it("FindAll", "should allow retrieving all entries for a processor", func(t *testing.T) {
		r := new()
		err := r.Add("unit-tester", ArtifactRegistryEntry{
			ArtifactID: "artifact_one",
		})
		require.NoError(t, err)
		err = r.Add("unit-tester", ArtifactRegistryEntry{
			ArtifactID: "artifact_two",
		})
		require.NoError(t, err)

		all, err := r.FindAll("unit-tester")
		require.NoError(t, err)
		assert.ElementsMatch(t, []ArtifactRegistryEntry{
			{ArtifactID: "artifact_one"},
			{ArtifactID: "artifact_two"},
		}, all)
	})

	it("FindAll", "should return nil for a non-existent processor", func(t *testing.T) {
		r := new()

		all, err := r.FindAll("nonexistent-tester")
		require.NoError(t, err)
		assert.Nil(t, all)
	})

	it("FindAll", "should return correct entries after adding and removing some", func(t *testing.T) {
		r := new()
		err := r.Add("unit-tester", ArtifactRegistryEntry{
			ArtifactID: "artifact_one",
		})
		require.NoError(t, err)
		err = r.Add("unit-tester", ArtifactRegistryEntry{
			ArtifactID: "artifact_two",
		})
		require.NoError(t, err)

		err = r.Remove("unit-tester", "artifact_one")
		require.NoError(t, err)

		all, err := r.FindAll("unit-tester")
		require.NoError(t, err)
		assert.Equal(t, []ArtifactRegistryEntry{
			{ArtifactID: "artifact_two"},
		}, all)
	})

	// FindByID
	it("FindByID", "should find an entry by its ID", func(t *testing.T) {
		r := new()
		err := r.Add("unit-tester", ArtifactRegistryEntry{
			ArtifactID: "artifact_to_find",
		})
		require.NoError(t, err)

		entry, found, err := r.FindByID("unit-tester", "artifact_to_find")
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, ArtifactRegistryEntry{
			ArtifactID: "artifact_to_find",
		}, entry)
	})

	it("FindByID", "should return not found and no error when finding a non-existent entry by ID", func(t *testing.T) {
		r := new()

		entry, found, err := r.FindByID("unit-tester", "nonexistent_artifact")
		require.NoError(t, err)
		assert.False(t, found)
		assert.Equal(t, ArtifactRegistryEntry{}, entry)
	})

	// Save
	it("Save", "should save an empty registry without errors", func(t *testing.T) {
		r := new()

		err := r.Save()
		require.NoError(t, err)
	})

	it("Save", "should save a populated registry without errors", func(t *testing.T) {
		r := new()
		err := r.Add("unit-tester", ArtifactRegistryEntry{
			ArtifactID: "artifact_to_save",
		})
		require.NoError(t, err)

		err = r.Save()
		require.NoError(t, err)
	})
}
