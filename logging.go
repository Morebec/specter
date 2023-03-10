package specter

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"io"
	"sync"
)

// Logger interface to be used by specter and processors to perform logging.
// implementations can be made for different scenarios, such as outputting to a file, stderr, silencing the logger etc.
// The logger only provides contextual logging.
type Logger interface {
	// Trace should only be used for debugging purposes.
	Trace(msg string)

	// Info is used to indicate informative messages.
	Info(msg string)

	// Warning is used to indicate events that could be problematic but that do not constitute errors.
	Warning(msg string)

	// Error is used to indicate that an error has occurred within specter.
	Error(msg string)

	// Success is used to indicate that a given action was performed successfully.
	// This can be used in stdout for example to format specific colors for successful actions as opposed to Info.
	Success(msg string)
}

type ColoredOutputLoggerConfig struct {
	EnableColors bool
	Writer       io.Writer
}

type ColoredOutputLogger struct {
	color  aurora.Aurora
	writer io.Writer
	mux    sync.Mutex
}

func NewColoredOutputLogger(c ColoredOutputLoggerConfig) *ColoredOutputLogger {
	return &ColoredOutputLogger{
		color:  aurora.NewAurora(c.EnableColors),
		writer: c.Writer,
	}
}

func (l ColoredOutputLogger) Trace(msg string) {
	l.Log(l.color.Faint(msg).String())
}

func (l ColoredOutputLogger) Info(msg string) {
	l.Log(msg)
}

func (l ColoredOutputLogger) Warning(msg string) {
	l.Log(l.color.Bold(l.color.Yellow(msg)).String())
}

func (l ColoredOutputLogger) Error(msg string) {
	l.Log(l.color.Bold(l.color.Red(msg)).String())
}

func (l ColoredOutputLogger) Success(msg string) {
	l.Log(l.color.Green(msg).String())
}

func (l ColoredOutputLogger) Log(msg string) {
	defer l.mux.Unlock()
	l.mux.Lock()
	fmt.Fprintln(l.writer, msg)
}
