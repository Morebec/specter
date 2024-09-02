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

package specter_test

import (
	"bytes"
	"github.com/logrusorgru/aurora"
	. "github.com/morebec/specter/pkg/specter"
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
				require.Equal(t, os.Stdout, l.Writer)

				// Check colors enabled by capturing artifact
				buffer := bytes.Buffer{}
				l.Writer = &buffer
				l.Success("hello world")

				assert.Equal(t, aurora.Green("hello world").String()+"\n", buffer.String())
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
				l.Writer = &buffer
				l.Success("hello world")

				assert.Equal(t, "hello world\n", buffer.String())
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
	assert.Equal(t, aurora.Faint("--- hello world").String()+"\n", buffer.String())
}

func TestDefaultLogger_Info(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewDefaultLogger(DefaultLoggerConfig{
		Writer: buffer,
	})

	logger.Info("hello world")
	assert.Equal(t, "hello world\n", buffer.String())
}

func TestDefaultLogger_Warning(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewDefaultLogger(DefaultLoggerConfig{
		Writer: buffer,
	})

	logger.Warning("hello world")
	assert.Equal(t, aurora.Bold(aurora.Yellow("hello world")).String()+"\n", buffer.String())
}

func TestDefaultLogger_Success(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewDefaultLogger(DefaultLoggerConfig{
		Writer: buffer,
	})

	logger.Success("hello world")
	assert.Equal(t, aurora.Green("hello world").String()+"\n", buffer.String())
}

func TestDefaultLogger_Error(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewDefaultLogger(DefaultLoggerConfig{
		Writer: buffer,
	})

	logger.Error("hello world")
	assert.Equal(t, aurora.Bold(aurora.Red("hello world")).String()+"\n", buffer.String())
}
