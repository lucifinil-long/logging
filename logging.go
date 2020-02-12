package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	loggers map[string]Logger
	rwlock  sync.RWMutex
)

func init() {
	loggers = make(map[string]Logger, 10)
}

// GetLogger will get a logger object that component, logname specified;
//	will using logDir, backupDir, level to create a new logger object if not found
func GetLogger(logDir, backupDir string,
	level LogLevel,
	component, logname string, suffixes ...string) (Logger, error) {
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

	logger, err := newLogger(logDir, backupDir, level, component, logname, suffix)
	if err != nil {
		return nil, err
	}
	loggers[key] = logger

	return logger, nil
}

func newLogger(logDir, backupDir string,
	level LogLevel,
	component, logname, suffix string) (Logger, error) {
	// try to create log directory
	os.MkdirAll(logDir, os.ModePerm)
	os.MkdirAll(backupDir, os.ModePerm)

	filename := filepath.Join(logDir, fmt.Sprintf("%s-%s", component, logname))
	logger, err := buildLogger(filename, suffix, backupDir)
	if err != nil {
		fmt.Println("build logger failed with", err)
		return nil, err
	}

	fmt.Println("log file suffix is", filename)
	logger.SetLevel(level)

	return logger, nil
}
