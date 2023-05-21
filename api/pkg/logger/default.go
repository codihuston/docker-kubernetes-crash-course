package logger

import "fmt"

// DefaultLogger is a simple logger that prints all messages to STDOUT with a
// prefix of either of the following using `fmt`: TRACE, DEBUG, INFO, WARN,
// ERROR, FATAL.
type DefaultLogger struct {
	Logger
}

func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{}
}

// unpack returns a string with spaces between each word.
func unpack(s ...any) string {
	var result string
	for i, w := range s {
		// Do append space after final word.
		if i == len(s)-1 {
			result += fmt.Sprint(w)
		} else {
			result += fmt.Sprint(w, " ")
		}
	}
	return result
}

func (l *DefaultLogger) Config() {
	// no-op.
}

func (l *DefaultLogger) Trace(s ...any) {
	fmt.Println("TRACE:", unpack(s...))
}

func (l *DefaultLogger) Debug(s ...any) {
	fmt.Println("DEBUG:", unpack(s...))
}

func (l *DefaultLogger) Info(s ...any) {
	fmt.Println("INFO:", unpack(s...))
}

func (l *DefaultLogger) Warn(s ...any) {
	fmt.Println("WARN:", unpack(s...))
}

func (l *DefaultLogger) Error(s ...any) {
	fmt.Println("ERROR:", unpack(s...))
}

func (l *DefaultLogger) Fatal(s ...any) {
	fmt.Println("FATAL:", unpack(s...))
}
