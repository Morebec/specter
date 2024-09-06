package specter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSourceLoadingStageHooksAdapter(t *testing.T) {
	t.Run("TestSourceLoadingStageHooksAdapter_Before", func(t *testing.T) {
		a := SourceLoadingStageHooksAdapter{}
		err := a.Before(PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("TestSourceLoadingStageHooksAdapter_After", func(t *testing.T) {
		a := SourceLoadingStageHooksAdapter{}
		err := a.After(PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("TestSourceLoadingStageHooksAdapter_BeforeSourceLocation", func(t *testing.T) {
		a := SourceLoadingStageHooksAdapter{}
		err := a.BeforeSourceLocation(PipelineContext{}, "")
		require.NoError(t, err)
	})

	t.Run("TestSourceLoadingStageHooksAdapter_AfterSourceLocation", func(t *testing.T) {
		a := SourceLoadingStageHooksAdapter{}
		err := a.AfterSourceLocation(PipelineContext{}, "")
		require.NoError(t, err)
	})

	t.Run("TestSourceLoadingStageHooksAdapter_OnError", func(t *testing.T) {
		a := SourceLoadingStageHooksAdapter{}
		err := a.OnError(PipelineContext{}, assert.AnError)
		require.Equal(t, assert.AnError, err)
	})
}

func TestUnitLoadingStageHooksAdapter(t *testing.T) {
	t.Run("TestUnitLoadingStageHooksAdapter_Before", func(t *testing.T) {

		a := UnitLoadingStageHooksAdapter{}
		err := a.Before(PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("TestUnitLoadingStageHooksAdapter_After", func(t *testing.T) {

		a := UnitLoadingStageHooksAdapter{}
		err := a.After(PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("TestUnitLoadingStageHooksAdapter_BeforeSource", func(t *testing.T) {

		a := UnitLoadingStageHooksAdapter{}
		err := a.BeforeSource(PipelineContext{}, Source{})
		require.NoError(t, err)
	})

	t.Run("TestUnitLoadingStageHooksAdapter_AfterSource", func(t *testing.T) {

		a := UnitLoadingStageHooksAdapter{}
		err := a.AfterSource(PipelineContext{}, Source{})
		require.NoError(t, err)
	})

	t.Run("TestUnitLoadingStageHooksAdapter_OnError", func(t *testing.T) {

		a := UnitLoadingStageHooksAdapter{}
		err := a.OnError(PipelineContext{}, assert.AnError)
		require.Equal(t, assert.AnError, err)
	})
}

func TestUnitProcessingStageHooksAdapter(t *testing.T) {
	t.Run("TestUnitProcessingStageHooksAdapter_Before", func(t *testing.T) {
		a := UnitProcessingStageHooksAdapter{}
		err := a.Before(PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("TestUnitProcessingStageHooksAdapter_After", func(t *testing.T) {
		a := UnitProcessingStageHooksAdapter{}
		err := a.After(PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("TestUnitProcessingStageHooksAdapter_BeforeProcessor", func(t *testing.T) {
		a := UnitProcessingStageHooksAdapter{}
		err := a.BeforeProcessor(PipelineContext{}, "processor")
		require.NoError(t, err)
	})

	t.Run("TestUnitProcessingStageHooksAdapter_AfterProcessor", func(t *testing.T) {
		a := UnitProcessingStageHooksAdapter{}
		err := a.AfterProcessor(PipelineContext{}, "processor")
		require.NoError(t, err)
	})

	t.Run("TestUnitProcessingStageHooksAdapter_OnError", func(t *testing.T) {
		a := UnitProcessingStageHooksAdapter{}
		err := a.OnError(PipelineContext{}, assert.AnError)
		require.Equal(t, assert.AnError, err)
	})
}

func TestArtifactProcessingStageHooksAdapter(t *testing.T) {
	t.Run("TestArtifactProcessingStageHooksAdapter_Before", func(t *testing.T) {
		a := ArtifactProcessingStageHooksAdapter{}
		err := a.Before(PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("TestArtifactProcessingStageHooksAdapter_After", func(t *testing.T) {
		a := ArtifactProcessingStageHooksAdapter{}
		err := a.After(PipelineContext{})
		require.NoError(t, err)
	})

	t.Run("TestArtifactProcessingStageHooksAdapter_BeforeProcessor", func(t *testing.T) {
		a := ArtifactProcessingStageHooksAdapter{}
		err := a.BeforeProcessor(PipelineContext{}, "processor")
		require.NoError(t, err)
	})

	t.Run("TestArtifactProcessingStageHooksAdapter_AfterProcessor", func(t *testing.T) {
		a := ArtifactProcessingStageHooksAdapter{}
		err := a.AfterProcessor(PipelineContext{}, "processor")
		require.NoError(t, err)
	})

	t.Run("TestArtifactProcessingStageHooksAdapter_OnError", func(t *testing.T) {
		a := ArtifactProcessingStageHooksAdapter{}
		err := a.OnError(PipelineContext{}, assert.AnError)
		require.Equal(t, assert.AnError, err)
	})
}
