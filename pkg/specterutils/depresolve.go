package specterutils

// Copyright 2022 Morébec
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

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"github.com/morebec/specter/pkg/specter"
	"strings"
)

const DependencyResolutionFailed = "specter.dependency_resolution_failed"

const ResolvedDependenciesArtifactID = "_resolved_dependencies"

func GetResolvedDependenciesFromContext(ctx specter.UnitProcessingContext) ResolvedDependencies {
	return specter.GetContextArtifact[ResolvedDependencies](ctx, ResolvedDependenciesArtifactID)
}

// ResolvedDependencies represents an ordered list of Unit that should be
// processed in that specific order based on their dependencies.
type ResolvedDependencies specter.UnitGroup

func (r ResolvedDependencies) ID() specter.ArtifactID {
	return ResolvedDependenciesArtifactID
}

type DependencyProvider interface {
	Supports(s specter.Unit) bool
	Provide(s specter.Unit) []specter.UnitID
}

var _ specter.UnitProcessor = DependencyResolutionProcessor{}

// DependencyResolutionProcessor is both a specter.UnitPreprocessor and a specter.UnitProcessor that resolves the dependencies
// of units based on specific DependencyProvider.
// When used as a specter.UnitPreprocessor units will be topologically sorted as a result in the pipeline.
// When used as a specter.UnitProcessor the ResolvedDependencies will be available as a specter.Artifact under the key ResolvedDependenciesArtifactID.
// A helper function GetResolvedDependenciesFromContext can be used in other processors to get access to the artifacts.
type DependencyResolutionProcessor struct {
	providers []DependencyProvider
}

func NewDependencyResolutionProcessor(providers ...DependencyProvider) *DependencyResolutionProcessor {
	return &DependencyResolutionProcessor{providers: providers}
}

func (p DependencyResolutionProcessor) Name() string {
	return "dependency_resolution_processor"
}

func (p DependencyResolutionProcessor) Preprocess(context specter.PipelineContext, units []specter.Unit) ([]specter.Unit, error) {
	deps, err := p.resolveDependencies(units)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func (p DependencyResolutionProcessor) Process(ctx specter.UnitProcessingContext) ([]specter.Artifact, error) {
	deps, err := p.resolveDependencies(ctx.Units)
	if err != nil {
		return nil, err
	}
	return []specter.Artifact{deps}, nil
}

func (p DependencyResolutionProcessor) resolveDependencies(units []specter.Unit) (ResolvedDependencies, error) {
	var nodes []dependencyNode
	for _, u := range units {
		node := dependencyNode{Unit: u, Dependencies: nil}
		for _, provider := range p.providers {
			if !provider.Supports(u) {
				continue
			}
			deps := provider.Provide(u)
			node.Dependencies = newDependencySet(deps...)
			break
		}
		nodes = append(nodes, node)
	}

	deps, err := newDependencyGraph(nodes...).resolve()
	if err != nil {
		return nil, errors.WrapWithMessage(err, DependencyResolutionFailed, "failed resolving dependencies")
	}
	return deps, nil
}

type dependencySet map[specter.UnitID]struct{}

func newDependencySet(dependencies ...specter.UnitID) dependencySet {
	deps := dependencySet{}
	for _, d := range dependencies {
		deps[d] = struct{}{}
	}

	return deps
}

// diff Returns all elements that are in s but not in o. A / B
func (s dependencySet) diff(o dependencySet) dependencySet {
	diff := dependencySet{}

	for d := range s {
		if _, found := o[d]; !found {
			diff[d] = s[d]
		}
	}

	return diff
}

type dependencyNode struct {
	Unit         specter.Unit
	Dependencies dependencySet
}

func (d dependencyNode) UnitName() specter.UnitID {
	return d.Unit.ID()
}

type dependencyGraph []dependencyNode

func newDependencyGraph(units ...dependencyNode) dependencyGraph {
	return append(dependencyGraph{}, units...)
}

func (g dependencyGraph) resolve() (ResolvedDependencies, error) {
	var resolved ResolvedDependencies

	// Look up of nodes to their IDs.
	unitByID := map[specter.UnitID]specter.Unit{}

	// Map nodes to dependencies
	dependenciesByID := map[specter.UnitID]dependencySet{}
	for _, n := range g {
		unitByID[n.UnitName()] = n.Unit
		dependenciesByID[n.UnitName()] = n.Dependencies
	}

	// The algorithm simply processes all nodes and tries to find the ones that have
	// no dependencies. When a node has dependencies, these dependencies are checked
	// for being either circular or unresolvable. If no unresolvable or circular
	// dependency is found, the node is considered resolved. And processing retries
	// with the remaining dependent nodes.
	for len(dependenciesByID) != 0 {
		var idsWithNoDependencies []specter.UnitID
		for id, dependencies := range dependenciesByID {
			if len(dependencies) == 0 {
				idsWithNoDependencies = append(idsWithNoDependencies, id)
			}
		}

		// If no nodes have no dependencies, in other words if all nodes have dependencies,
		// This means that we have a problem of circular dependencies.
		// We need at least one node in the graph to be independent for it to be potentially resolvable.
		if len(idsWithNoDependencies) == 0 {
			// We either have circular dependencies or an unresolved dependency
			// Check if all dependencies exist.
			for id, dependencies := range dependenciesByID {
				for dependency := range dependencies {
					if _, found := unitByID[dependency]; !found {
						return nil, errors.NewWithMessage(
							DependencyResolutionFailed,
							fmt.Sprintf("unit %q depends on an unresolved kind %q",
								id,
								dependency,
							),
						)
					}
				}
			}

			// They all exist, therefore, we have a circular dependencies.
			var circularDependencies []string
			for k := range dependenciesByID {
				circularDependencies = append(circularDependencies, string(k))
			}

			return nil, errors.NewWithMessage(
				DependencyResolutionFailed,
				fmt.Sprintf(
					"circular dependencies found between nodes %q",
					strings.Join(circularDependencies, "\", \""),
				),
			)
		}

		// All good, we can move the nodes that no longer have unresolved dependencies
		for _, nodeID := range idsWithNoDependencies {
			delete(dependenciesByID, nodeID)
			resolved = append(resolved, unitByID[nodeID])
		}

		// Remove the resolved nodes from the remaining dependenciesByID.
		for ID, dependencies := range dependenciesByID {
			diff := dependencies.diff(newDependencySet(idsWithNoDependencies...))
			dependenciesByID[ID] = diff
		}
	}

	return resolved, nil
}

// HasDependencies is an interface that can be implemented by units
// that define their dependencies from their field values.
// This interface can be used in conjunction with the HasDependenciesProvider
// to easily resolve dependencies.
type HasDependencies interface {
	specter.Unit
	Dependencies() []specter.UnitID
}

type HasDependenciesProvider struct{}

func (h HasDependenciesProvider) Supports(u specter.Unit) bool {
	_, ok := u.(HasDependencies)
	return ok
}

func (h HasDependenciesProvider) Provide(u specter.Unit) []specter.UnitID {
	d, ok := u.(HasDependencies)
	if !ok {
		return nil
	}
	return d.Dependencies()
}
