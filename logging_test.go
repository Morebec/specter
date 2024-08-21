package specter

import (
	"bytes"
	"github.com/logrusorgru/aurora"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNewDefaultLogger(t *testing.T) {
	buffer := &bytes.Buffer{}
	tests := []struct {
		name  string
		given DefaultLoggerConfig
		want  func(*DefaultLogger)
	}{
		{
			name:  "GIVEN a zero value config, THEN have a logger with os.Stdout as a writer and colors enabled",
			given: DefaultLoggerConfig{},
			want: func(l *DefaultLogger) {
				require.NotNil(t, l)
				require.Equal(t, os.Stdout, l.writer)

				// Check colors enabled by capturing output
				buffer := bytes.Buffer{}
				l.writer = &buffer
				l.Success("hello world")

				assert.Equal(t, aurora.Green("hello world").String()+"\n", string(buffer.Bytes()))
			},
		},
		{
			name: "GIVEN a config with color disabled, THEN have a logger with no colors enabled",
			given: DefaultLoggerConfig{
				DisableColors: true,
			},
			want: func(l *DefaultLogger) {
				// Check colors enabled
				buffer := bytes.Buffer{}
				l.writer = &buffer
				l.Success("hello world")

				assert.Equal(t, "hello world\n", string(buffer.Bytes()))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer.Reset()
			logger := NewDefaultLogger(tt.given)
			tt.want(logger)
		})
	}
}

func TestDefaultLogger_Trace(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewDefaultLogger(DefaultLoggerConfig{
		Writer: buffer,
	})

	logger.Trace("hello world")
	assert.Equal(t, aurora.Faint("hello world").String()+"\n", string(buffer.Bytes()))
}

func TestDefaultLogger_Info(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewDefaultLogger(DefaultLoggerConfig{
		Writer: buffer,
	})

	logger.Info("hello world")
	assert.Equal(t, "hello world\n", string(buffer.Bytes()))
}

func TestDefaultLogger_Warning(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewDefaultLogger(DefaultLoggerConfig{
		Writer: buffer,
	})

	logger.Warning("hello world")
	assert.Equal(t, aurora.Bold(aurora.Yellow("hello world")).String()+"\n", string(buffer.Bytes()))
}

func TestDefaultLogger_Success(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewDefaultLogger(DefaultLoggerConfig{
		Writer: buffer,
	})

	logger.Success("hello world")
	assert.Equal(t, aurora.Green("hello world").String()+"\n", string(buffer.Bytes()))
}

func TestDefaultLogger_Error(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewDefaultLogger(DefaultLoggerConfig{
		Writer: buffer,
	})

	logger.Error("hello world")
	assert.Equal(t, aurora.Bold(aurora.Red("hello world")).String()+"\n", string(buffer.Bytes()))
}
