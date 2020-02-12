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

const (
	dateFormat = "2006-01-02"
	hourFormat = "2006010215"
)

// Config stores the config for logger
type Config struct {
	// LogDir stores log dir
	LogDir string
	// BackupDir stores backup dir
	BackupDir string
	// Level describe log level
	Level LogLevel
}

// DefaultConfig stores a default logging config
var DefaultConfig = &Config{
	LogDir:    "./logs/now",
	BackupDir: "./logs/backups",
	Level:     LogLevelTrace,
}
