package logger

// NoopLogger is a logger that discards all output. Useful for tests
// or anywhere logging should be silenced.
type NoopLogger struct{}

// NewNoop creates a no-op logger that satisfies the Logger interface.
func NewNoop() Logger {
	return &NoopLogger{}
}

func (n *NoopLogger) Info(string)          {}
func (n *NoopLogger) Event(string)         {}
func (n *NoopLogger) Debug(string, string) {}
func (n *NoopLogger) Error(error, string)  {}
func (n *NoopLogger) Scope(string) Logger  { return n }
