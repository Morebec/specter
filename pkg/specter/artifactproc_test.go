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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestArtifactProcessorFunc(t *testing.T) {
	t.Run("Name should be set", func(t *testing.T) {
		a := NewArtifactProcessorFunc("name", func(ctx ArtifactProcessingContext) error {
			return nil
		})
		require.Equal(t, "name", a.Name())
	})

	t.Run("Process should be called", func(t *testing.T) {
		called := false
		a := NewArtifactProcessorFunc("name", func(ctx ArtifactProcessingContext) error {
			called = true
			return assert.AnError
		})

		err := a.Process(ArtifactProcessingContext{})
		require.Error(t, err)
		require.True(t, called)
	})
}
