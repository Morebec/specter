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
	"context"
)

// ArtifactProcessor are services responsible for processing artifacts of UnitProcessors.
type ArtifactProcessor interface {
	// Process performs the processing of artifacts generated by UnitProcessor.
	Process(ctx ArtifactProcessingContext) error

	// Name returns the name of this processor.
	Name() string
}

// ArtifactRegistry provides an interface for managing a registry of artifacts. This
// registry tracks artifacts generated during processing runs, enabling clean-up
// in subsequent runs to avoid residual artifacts and maintain a clean slate.
//
// Implementations of the ArtifactRegistry interface must be thread-safe to handle
// concurrent calls to TrackFile and UntrackFile methods. Multiple goroutines may
// access the registry simultaneously, so appropriate synchronization mechanisms
// should be implemented to prevent race conditions and ensure data integrity.
type ArtifactRegistry interface {
	// Load the registry state from persistent storage. If an error occurs, it
	// should be returned to indicate the failure of the loading operation.
	Load() error

	// Save the current state of the registry to persistent storage. If an
	// error occurs, it should be returned to indicate the failure of the saving operation.
	Save() error

	// Add registers an ArtifactID under a specific processor name. This method
	// should ensure that the file path is associated with the given processor name
	// in the registry.
	Add(processorName string, e ArtifactRegistryEntry) error

	// Remove a given ArtifactID artifact registration for a specific processor name. This
	// method should ensure that the file path is disassociated from the given
	// processor name in the registry.
	Remove(processorName string, artifactID ArtifactID) error

	// FindByID finds an entry by its ArtifactID.
	FindByID(processorName string, artifactID ArtifactID) (entry ArtifactRegistryEntry, found bool, err error)

	// FindAll returns all the entries in the registry.
	FindAll(processorName string) ([]ArtifactRegistryEntry, error)
}

type ArtifactRegistryEntry struct {
	ArtifactID ArtifactID
	Metadata   map[string]any
}

// ProcessorArtifactRegistry is a wrapper around an ArtifactRegistry that scopes all calls to a given processor.
type ProcessorArtifactRegistry struct {
	processorName string
	registry      ArtifactRegistry
}

func NewProcessorArtifactRegistry(processorName string, registry ArtifactRegistry) ProcessorArtifactRegistry {
	return ProcessorArtifactRegistry{processorName: processorName, registry: registry}
}

func (n ProcessorArtifactRegistry) Add(artifactID ArtifactID, metadata map[string]any) error {
	return n.registry.Add(n.processorName, ArtifactRegistryEntry{
		ArtifactID: artifactID,
		Metadata:   metadata,
	})
}

func (n ProcessorArtifactRegistry) Remove(artifactID ArtifactID) error {
	return n.registry.Remove(n.processorName, artifactID)
}

func (n ProcessorArtifactRegistry) FindByID(artifactID ArtifactID) (_ ArtifactRegistryEntry, _ bool, _ error) {
	return n.registry.FindByID(n.processorName, artifactID)
}

func (n ProcessorArtifactRegistry) FindAll() ([]ArtifactRegistryEntry, error) {
	return n.registry.FindAll(n.processorName)
}

type ArtifactProcessingContext struct {
	context.Context
	Units     UnitGroup
	Artifacts []Artifact

	ArtifactRegistry ProcessorArtifactRegistry
	processorName    string
}
