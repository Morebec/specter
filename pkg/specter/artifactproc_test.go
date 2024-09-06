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
