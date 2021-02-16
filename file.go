package logging

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	logDebug = "debug"
	logTrace = "trace"
	logWarn  = "warn "
	logError = "error"
)

const (
	_ = iota
	// KB describes KB of file size
	KB int64 = 1 << (iota * 10)
	// MB describes MB of file size
	MB
	// GB describes MB of file size
	GB
	// TB describes MB of file size
	TB
	maxFileSize       = 2 * GB
	maxFileCount      = 10
	defaultBufferSize = 2 * KB
)

var logLevels = []string{logDebug, logTrace, logWarn, logError}

type fileLogger struct {
	sync.RWMutex
	logInfoMap map[string]*loggerInfo
	suffixInfo string
	logLevel   uint8
}

func (l *fileLogger) String() string {
	ret := fmt.Sprintf("{\"level\":%v, \"suffix\":%v, \"log info map\":%v}",
		l.logLevel, l.suffixInfo, l.logInfoMap)
	return ret
}

func (l *fileLogger) Write(filename string, suffix bool, args ...interface{}) {
	l.Lock()
	defer l.Unlock()

	var err error
	loggerInfo, ok := l.logInfoMap[filename]
	if !ok { //不存在需要重新初始化一下
		// reinit logger info object if not exist
		if loggerInfo, err = newLoggerInfo(filename, ""); err != nil {
			println(time.Now().Format("2006-01-02 15:04:05.000:"), "[NewLoggerInfo] Write : "+err.Error())
			return
		}
		go loggerInfo.WriteBufferToQueue()
		go loggerInfo.FlushBufferQueue()
		l.logInfoMap[filename] = loggerInfo
	}
	loggerInfo.Write(format("Raw", suffix, l.suffixInfo, args...))
}

func (l *fileLogger) SetLevel(level uint8) {
	l.Lock()
	defer l.Unlock()
	if level > LogLevelError {
		l.logLevel = LogLevelTrace
	} else {
		l.logLevel = level
	}
}

func (l *fileLogger) CheckLevel(level uint8) bool {
	l.RLock()
	defer l.RUnlock()
	return l.logLevel <= level
}

func (l *fileLogger) Debug(args ...interface{}) {
	logs := addFuncNameToLogs(defaultDepth, args)

	l.debug(logs...)
}

func (l *fileLogger) debug(args ...interface{}) {
	l.RLock()
	loggerInfo := l.logInfoMap[logDebug]
	d := l.CheckLevel(LogLevelDebug)
	l.RUnlock()
	if !d {
		return
	}

	hasSuffix := true
	if l.suffixInfo == "" {
		hasSuffix = false
	}

	log := format(logDebug, hasSuffix, l.suffixInfo, args...)
	fmt.Print(log)
	loggerInfo.Write(log)
}

func (l *fileLogger) Trace(args ...interface{}) {
	logs := addFuncNameToLogs(defaultDepth, args)

	l.trace(logs...)
}

func (l *fileLogger) trace(args ...interface{}) {
	l.RLock()
	loggerInfo := l.logInfoMap[logTrace]
	d := l.CheckLevel(LogLevelTrace)
	l.RUnlock()
	if !d {
		return
	}

	l.debug(args...)

	hasSuffix := true
	if l.suffixInfo == "" {
		hasSuffix = false
	}
	loggerInfo.Write(format(logTrace, hasSuffix, l.suffixInfo, args...))
}

func (l *fileLogger) Warn(args ...interface{}) {
	logs := addFuncNameToLogs(defaultDepth, args)

	l.warn(logs...)
}

func (l *fileLogger) warn(args ...interface{}) {
	l.RLock()
	loggerInfo := l.logInfoMap[logWarn]
	d := l.CheckLevel(LogLevelWarn)
	l.RUnlock()
	if !d {
		return
	}

	l.trace(args...)

	hasSuffix := true
	if l.suffixInfo == "" {
		hasSuffix = false
	}
	loggerInfo.Write(format(logWarn, hasSuffix, l.suffixInfo, args...))
}

func (l *fileLogger) Error(args ...interface{}) {
	l.RLock()
	loggerInfo := l.logInfoMap[logError]
	d := l.CheckLevel(LogLevelError)
	l.RUnlock()
	if !d {
		return
	}

	logs := addFuncNameToLogs(defaultDepth, args)

	l.warn(logs...)

	hasSuffix := true
	if l.suffixInfo == "" {
		hasSuffix = false
	}
	loggerInfo.Write(format(logError, hasSuffix, l.suffixInfo, logs...))
}

func (l *fileLogger) Shutdown() error {
	for k, v := range l.logInfoMap {
		v.Close()
		println(time.Now().Format("2006-01-02 15:04:05.000:"), "closed", k, "logging.")
	}
	return nil
}

func buildLogger(filename, suffix, historyDir string) (Logger, error) {
	logInfoMap := make(map[string]*loggerInfo, 5)
	for _, level := range logLevels {
		logInfo, err := newLoggerInfo(filename, level)
		if err != nil {
			return nil, err
		}
		logInfo.historyDir = historyDir
		go logInfo.WriteBufferToQueue()
		go logInfo.FlushBufferQueue()
		logInfoMap[level] = logInfo
	}

	return &fileLogger{
		logInfoMap: logInfoMap,
		suffixInfo: suffix,
	}, nil
}

func getDatetime() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

func format(prefix string, suffix bool, suffixInfo string, args ...interface{}) string {
	content := "|" + prefix
	for _, arg := range args {
		switch arg.(type) {
		case int:
			content = content + "|" + strconv.Itoa(arg.(int))
			break
		case string:
			content = content + "|" + strings.TrimRight(arg.(string), "\n")
			break
		case int64:
			str := strconv.FormatInt(arg.(int64), 10)
			content = content + "|" + str
			break
		default:
			content = content + "|" + fmt.Sprintf("%v", arg)
			break
		}
	}
	if suffix {
		content = getDatetime() + content + "|" + suffixInfo + "\n"
	} else {
		content = getDatetime() + content + "\n"
	}
	return content
}
