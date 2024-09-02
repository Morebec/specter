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

import "context"

type ProcessingContext struct {
	context.Context
	Specifications SpecificationGroup
	Artifacts      []Artifact
	Logger         Logger
}

// Artifact returns the artifact associated with a given processor.
func (c ProcessingContext) Artifact(id ArtifactID) Artifact {
	for _, o := range c.Artifacts {
		if o.ID() == id {
			return o
		}
	}
	return nil
}

type ArtifactID string

// Artifact represents a result or output generated by a SpecificationProcessor.
// An artifact is a unit of data or information produced as part of the processing workflow.
// It can be a transient, in-memory object, or it might represent more permanent entities such as
// files on disk, records in a database, deployment units, or other forms of data artifacts.
type Artifact interface {
	// ID is a unique identifier of the artifact, which helps in distinguishing and referencing
	// the artifact within a processing context.
	ID() ArtifactID
}

// SpecificationProcessor are services responsible for performing work using Specifications
// and which can possibly generate artifacts.
type SpecificationProcessor interface {
	// Name returns the unique name of this processor.
	// This name can appear in logs to report information about a given processor.
	Name() string

	// Process processes a group of specifications.
	Process(ctx ProcessingContext) ([]Artifact, error)
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
	Specifications SpecificationGroup
	Artifacts      []Artifact
	Logger         Logger

	ArtifactRegistry ProcessorArtifactRegistry
	processorName    string
}

// ArtifactProcessor are services responsible for processing artifacts of SpecProcessors.
type ArtifactProcessor interface {
	// Process performs the processing of artifacts generated by SpecificationProcessor.
	Process(ctx ArtifactProcessingContext) error

	// Name returns the name of this processor.
	Name() string
}

func GetContextArtifact[T Artifact](ctx ProcessingContext, id ArtifactID) T {
	artifact := ctx.Artifact(id)
	v, _ := artifact.(T)
	return v
}