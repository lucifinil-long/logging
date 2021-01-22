package logging

const (
	// LogLevelDebug marks debug logging level
	LogLevelDebug uint8 = iota
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
	// HistoryDir stores backup dir
	HistoryDir string
	// Level describe log level
	Level uint8
}

var (
	// DefaultLogDir defines default log dir
	DefaultLogDir = "./logs/now"
	// DefaultHistoryDir defines default history log dir
	DefaultHistoryDir = "./logs/history"
)
