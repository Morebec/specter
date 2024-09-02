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
	"fmt"
	"github.com/logrusorgru/aurora"
	"io"
	"os"
	"sync"
)

// Logger interface to be used by a Pipeline and its processors to perform logging.
type Logger interface {
	// Trace should only be used for debugging purposes.
	Trace(msg string)

	// Info is used to indicate informative messages.
	Info(msg string)

	// Warning is used to indicate events that could be problematic but that do not constitute errors.
	Warning(msg string)

	// Error is used to indicate that an error has occurred.
	Error(msg string)

	// Success is used to indicate that a given action was performed successfully.
	// This can be used in stdout for example to format specific colors for successful actions as opposed to Info.
	Success(msg string)
}

type DefaultLoggerConfig struct {
	DisableColors bool
	Writer        io.Writer
}

type DefaultLogger struct {
	color  aurora.Aurora
	Writer io.Writer
	mux    sync.Mutex
}

func NewDefaultLogger(c DefaultLoggerConfig) *DefaultLogger {
	writer := c.Writer
	if writer == nil {
		writer = os.Stdout
	}

	return &DefaultLogger{
		color:  aurora.NewAurora(!c.DisableColors),
		Writer: writer,
	}
}

func (l *DefaultLogger) Trace(msg string) {
	l.Log(l.color.Faint(fmt.Sprintf("--- %s", msg)).String())
}

func (l *DefaultLogger) Info(msg string) {
	l.Log(msg)
}

func (l *DefaultLogger) Warning(msg string) {
	l.Log(l.color.Bold(l.color.Yellow(msg)).String())
}

func (l *DefaultLogger) Error(msg string) {
	l.Log(l.color.Bold(l.color.Red(msg)).String())
}

func (l *DefaultLogger) Success(msg string) {
	l.Log(l.color.Green(msg).String())
}

func (l *DefaultLogger) Log(msg string) {
	defer l.mux.Unlock()
	l.mux.Lock()
	_, _ = fmt.Fprintln(l.Writer, msg)
}
