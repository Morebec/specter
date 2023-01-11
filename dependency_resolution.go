package specter

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
	"strings"
)

type DependencySet map[SpecName]struct{}

func newDependencySetForSpec(s Spec) DependencySet {
	set := DependencySet{}
	for _, d := range s.Dependencies() {
		set[d] = struct{}{}
	}
	return set
}

// diff Returns all elements that are in s and not in o. A / B
func (s DependencySet) diff(o DependencySet) DependencySet {
	diff := DependencySet{}

	for d := range s {
		if _, found := o[d]; !found {
			diff[d] = s[d]
		}
	}

	return diff
}

func (s DependencySet) Names() []SpecName {
	var typeNames []SpecName

	for k := range s {
		typeNames = append(typeNames, k)
	}

	return typeNames
}

func NewDependencySet(dependencies ...SpecName) DependencySet {
	deps := DependencySet{}
	for _, d := range dependencies {
		deps[d] = struct{}{}
	}

	return deps
}

// DependencyProvider are functions responsible for providing the dependencies of a list of Spec as a list of Spec.
// Generally providers are specialized for a specific SpecType.
type DependencyProvider func(systemSpec Spec, specs SpecGroup) ([]Spec, error)

type DependencyGraph []Spec

// Merge Allows merging this dependency graph with another one and returns the result.
func (g DependencyGraph) Merge(o DependencyGraph) DependencyGraph {
	var lookup = make(map[SpecName]bool)
	var merge []Spec

	for _, node := range g {
		merge = append(merge, node)
		lookup[node.Name()] = true
	}

	for _, node := range o {
		if _, found := lookup[node.Name()]; found {
			continue
		}
		merge = append(merge, node)
	}

	return NewDependencyGraph(merge...)
}

func NewDependencyGraph(specs ...Spec) DependencyGraph {
	return append(DependencyGraph{}, specs...)
}

func (g DependencyGraph) Resolve() (ResolvedDependencies, error) {
	var resolved []Spec

	// Look up of nodes to their typeName Names.
	specByTypeNames := map[SpecName]Spec{}

	// Map nodes to dependencies
	dependenciesByTypeNames := map[SpecName]DependencySet{}
	for _, n := range g {
		specByTypeNames[n.Name()] = n
		dependenciesByTypeNames[n.Name()] = newDependencySetForSpec(n)
	}

	// The algorithm simply processes all nodes and tries to find the ones that have no dependencies.
	// When a node has dependencies, these dependencies are checked for being either circular or unresolvable.
	// If no unresolvable or circular dependency is found, the node is considered resolved.
	// And processing retries with the remaining dependent nodes.
	for len(dependenciesByTypeNames) != 0 {
		var typeNamesWithNoDependencies []SpecName
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
							fmt.Sprintf("spec with type FilePath \"%s\" depends on an unresolved type FilePath \"%s\"",
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
					"circular dependencies found between nodes \"%s\"",
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
			diff := dependencies.diff(NewDependencySet(typeNamesWithNoDependencies...))
			dependenciesByTypeNames[typeName] = diff
		}
	}

	return append(ResolvedDependencies{}, resolved...), nil
}

// ResolvedDependencies represents an ordered list of Spec that should be processed in that specific order to avoid
// unresolved types.
// TODO Remove spec group and add its methods here.
type ResolvedDependencies SpecGroup
