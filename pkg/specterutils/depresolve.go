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

const ResolvedDependenciesArtifactID = "_resolved_dependencies"

// ResolvedDependencies represents an ordered list of Unit that should be processed in that specific order to avoid
// unresolved types.
type ResolvedDependencies specter.UnitGroup

func (r ResolvedDependencies) ID() specter.ArtifactID {
	return ResolvedDependenciesArtifactID
}

type DependencyProvider interface {
	Supports(s specter.Unit) bool
	Provide(s specter.Unit) []specter.UnitName
}

var _ specter.UnitProcessor = DependencyResolutionProcessor{}

type DependencyResolutionProcessor struct {
	providers []DependencyProvider
}

func NewDependencyResolutionProcessor(providers ...DependencyProvider) *DependencyResolutionProcessor {
	return &DependencyResolutionProcessor{providers: providers}
}

func (p DependencyResolutionProcessor) Name() string {
	return "dependency_resolution_processor"
}

func (p DependencyResolutionProcessor) Process(ctx specter.ProcessingContext) ([]specter.Artifact, error) {
	ctx.Logger.Info("\nResolving dependencies...")

	var nodes []dependencyNode
	for _, s := range ctx.Units {
		node := dependencyNode{Unit: s, Dependencies: nil}
		for _, provider := range p.providers {
			if !provider.Supports(s) {
				continue
			}
			deps := provider.Provide(s)
			node.Dependencies = newDependencySet(deps...)
			break
		}
		nodes = append(nodes, node)
	}

	deps, err := newDependencyGraph(nodes...).resolve()
	if err != nil {
		return nil, errors.WrapWithMessage(err, errors.InternalErrorCode, "failed resolving dependencies")
	}
	ctx.Logger.Success("Dependencies resolved successfully.")

	return []specter.Artifact{deps}, nil
}

func GetResolvedDependenciesFromContext(ctx specter.ProcessingContext) ResolvedDependencies {
	return specter.GetContextArtifact[ResolvedDependencies](ctx, ResolvedDependenciesArtifactID)
}

type dependencySet map[specter.UnitName]struct{}

func newDependencySet(dependencies ...specter.UnitName) dependencySet {
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

func (d dependencyNode) UnitName() specter.UnitName {
	return d.Unit.Name()
}

type dependencyGraph []dependencyNode

func newDependencyGraph(units ...dependencyNode) dependencyGraph {
	return append(dependencyGraph{}, units...)
}

func (g dependencyGraph) resolve() (ResolvedDependencies, error) {
	var resolved ResolvedDependencies

	// Look up of nodes to their typeName Names.
	specByTypeNames := map[specter.UnitName]specter.Unit{}

	// Map nodes to dependencies
	dependenciesByTypeNames := map[specter.UnitName]dependencySet{}
	for _, n := range g {
		specByTypeNames[n.UnitName()] = n.Unit
		dependenciesByTypeNames[n.UnitName()] = n.Dependencies
	}

	// The algorithm simply processes all nodes and tries to find the ones that have no dependencies.
	// When a node has dependencies, these dependencies are checked for being either circular or unresolvable.
	// If no unresolvable or circular dependency is found, the node is considered resolved.
	// And processing retries with the remaining dependent nodes.
	for len(dependenciesByTypeNames) != 0 {
		var typeNamesWithNoDependencies []specter.UnitName
		for typeName, dependencies := range dependenciesByTypeNames {
			if len(dependencies) == 0 {
				typeNamesWithNoDependencies = append(typeNamesWithNoDependencies, typeName)
			}
		}

		// If no nodes have no dependencies, in other words if all nodes have dependencies,
		// This means that we have a problem of circular dependencies.
		// We need at least one node in the graph to be independent for it to be potentially resolvable.
		if len(typeNamesWithNoDependencies) == 0 {
			// We either have circular dependencies or an unresolved dependency
			// Check if all dependencies exist.
			for typeName, dependencies := range dependenciesByTypeNames {
				for dependency := range dependencies {
					if _, found := specByTypeNames[dependency]; !found {
						return nil, errors.NewWithMessage(
							errors.InternalErrorCode,
							fmt.Sprintf("unit with type %q depends on an unresolved type %q",
								typeName,
								dependency,
							),
						)
					}
				}
			}

			// They all exist, therefore, we have a circular dependencies.
			var circularDependencies []string
			for k := range dependenciesByTypeNames {
				circularDependencies = append(circularDependencies, string(k))
			}

			return nil, errors.NewWithMessage(
				errors.InternalErrorCode,
				fmt.Sprintf(
					"circular dependencies found between nodes %q",
					strings.Join(circularDependencies, "\", \""),
				),
			)
		}

		// All good, we can move the nodes that no longer have unresolved dependencies
		for _, nodeTypeName := range typeNamesWithNoDependencies {
			delete(dependenciesByTypeNames, nodeTypeName)
			resolved = append(resolved, specByTypeNames[nodeTypeName])
		}

		// Remove the resolved nodes from the remaining dependenciesByTypeNames.
		for typeName, dependencies := range dependenciesByTypeNames {
			diff := dependencies.diff(newDependencySet(typeNamesWithNoDependencies...))
			dependenciesByTypeNames[typeName] = diff
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
	Dependencies() []specter.UnitName
}

type HasDependenciesProvider struct{}

func (h HasDependenciesProvider) Supports(u specter.Unit) bool {
	_, ok := u.(HasDependencies)
	return ok
}

func (h HasDependenciesProvider) Provide(u specter.Unit) []specter.UnitName {
	d, ok := u.(HasDependencies)
	if !ok {
		return nil
	}
	return d.Dependencies()
}
