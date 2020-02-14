package logging

// Logger defines logger interface
type Logger interface {
	// SetLevel
	SetLevel(uint8)
	CheckLevel(uint8) bool
	Write(string, bool, ...interface{})
	Debug(...interface{})
	Trace(...interface{})
	Warn(...interface{})
	Error(...interface{})
}
