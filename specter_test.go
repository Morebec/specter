package specter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithOutputRegistry(t *testing.T) {
	s := &Specter{}
	r := &JSONOutputRegistry{}
	WithOutputRegistry(r)(s)
	assert.Equal(t, r, s.OutputRegistry)
}

func TestWithDefaultLogger(t *testing.T) {
	s := &Specter{}
	WithDefaultLogger()(s)
	assert.IsType(t, &DefaultLogger{}, s.Logger)
}
