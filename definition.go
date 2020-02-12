package logging

// LogLevel describes logging level
type LogLevel uint32

const (
	// LogLevelDebug marks debug logging level
	LogLevelDebug LogLevel = iota
	// LogLevelTrace marks trace logging level
	LogLevelTrace
	// LogLevelWarn marks warn logging level
	LogLevelWarn
	// LogLevelError marks error logging level
	LogLevelError
)

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

const (
	dateFormat = "2006-01-02"
	hourFormat = "2006010215"
)
