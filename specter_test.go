package specter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithArtifactRegistry(t *testing.T) {
	s := &Specter{}
	r := &JSONArtifactRegistry{}
	WithArtifactRegistry(r)(s)
	assert.Equal(t, r, s.ArtifactRegistry)
}

func TestWithDefaultLogger(t *testing.T) {
	s := &Specter{}
	WithDefaultLogger()(s)
	assert.IsType(t, &DefaultLogger{}, s.Logger)
}
