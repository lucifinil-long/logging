package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	loggers map[string]Logger
	rwlock  sync.RWMutex
)

func init() {
	loggers = make(map[string]Logger, 10)
}

// DefaultConfig builds a default logging config
func DefaultConfig() *Config {
	return &Config{
		LogDir:     DefaultLogDir,
		HistoryDir: DefaultHistoryDir,
		Level:      LogLevelTrace,
	}
}

// StdoutLogger returns an stdout logger
func StdoutLogger(level uint8) Logger {
	if level > LogLevelError {
		level = LogLevelTrace
	}

	return &stdoutLogger{
		level: level,
	}
}

// FileLogger get a file logger object that component, logname specified;
//	will using config to create a new logger object if not found
func FileLogger(component, logname string, config *Config,
	suffixes ...string) (Logger, error) {
	if config == nil || config.LogDir == "" {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05.999:"), "logging.GetLogger",
			"invalid logging config, using default logging config")
		config = DefaultConfig()
	}

	key := fmt.Sprintf("%s-%s", component, logname)
	rwlock.RLock()
	logger, ok := loggers[key]
	rwlock.RUnlock()
	if ok {
		return logger, nil
	}

	rwlock.Lock()
	defer rwlock.Unlock()
	logger, ok = loggers[key]
	if ok {
		return logger, nil
	}

	suffix := ""
	if len(suffixes) > 0 {
		suffix = suffixes[0]
	}

	logger, err := newLogger(config.LogDir, config.HistoryDir, config.Level, component, logname, suffix)
	if err != nil {
		return nil, err
	}
	loggers[key] = logger

	return logger, nil
}

// GetLogger will get a logger
// 	if config is nil or log directory is not set, a screen logger is returned;
// 	otherwise a file logger is returned
func GetLogger(component, logname string, config *Config,
	suffixes ...string) (Logger, error) {
	if config == nil || config.LogDir == "" {
		level := LogLevelTrace
		if config != nil {
			level = config.Level
		}

		return StdoutLogger(level), nil
	}

	return FileLogger(component, logname, config, suffixes...)
}

func newLogger(logDir, historyDir string,
	level uint8,
	component, logname, suffix string) (Logger, error) {
	// try to create log directory
	os.MkdirAll(logDir, os.ModePerm)
	os.MkdirAll(historyDir, os.ModePerm)

	filename := filepath.Join(logDir, fmt.Sprintf("%s-%s", component, logname))
	logger, err := buildLogger(filename, suffix, historyDir)
	if err != nil {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05.999:"), "logging.GetLogger",
			"build logger failed with", err)
		return nil, err
	}

	logLevel := logLevels[int(LogLevelTrace)]
	if level >= LogLevelDebug && level <= LogLevelError {
		logLevel = logLevels[int(level)]
	}
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.999:"), "logging.GetLogger",
		"log file suffix is", filename, "while log level is", logLevel)
	logger.SetLevel(level)

	return logger, nil
}

var (
	defaultDepth = 3
)

func addFuncNameToLogs(depth int, args []interface{}) []interface{} {
	pc := make([]uintptr, 1)
	runtime.Callers(depth, pc)
	f := runtime.FuncForPC(pc[0])
	logs := make([]interface{}, 0, len(args)+1)
	logs = append(logs, f.Name())
	logs = append(logs, args...)

	return logs
}
