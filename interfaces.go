package logging

// Logger defines logger interface
type Logger interface {
	// SetLevel
	SetLevel(LogLevel)
	CheckLevel(LogLevel) bool
	Write(string, bool, ...interface{})
	Debug(...interface{})
	Trace(...interface{})
	Warn(...interface{})
	Error(...interface{})
}
