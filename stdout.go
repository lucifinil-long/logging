package logging

import (
	"fmt"
	"strings"
)

type stdoutLogger struct {
	level uint8
}

func (l *stdoutLogger) SetLevel(lv uint8) {
	if lv > LogLevelError {
		lv = LogLevelError
	}

	l.level = lv
}

func (l stdoutLogger) CheckLevel(lv uint8) bool {
	return l.level <= lv
}

func (l stdoutLogger) Write(suffixInfo string, suffix bool, args ...interface{}) {
	contents := format("", suffix, suffixInfo, args...)
	contents = strings.TrimSuffix(contents, "\n")
	fmt.Println(contents)
}

var (
	levels = []string{logDebug, logTrace, logWarn, logError}
)

func (l stdoutLogger) write(level uint8, args ...interface{}) {
	if l.CheckLevel(level) {
		logs := addFuncNameToLogs(defaultDepth+1, args)
		contents := format(levels[int(level)], false, "", logs...)
		contents = strings.TrimSuffix(contents, "\n")
		fmt.Println(contents)
	}
}

func (l stdoutLogger) Debug(args ...interface{}) {
	l.write(LogLevelDebug, args...)
}

func (l stdoutLogger) Trace(args ...interface{}) {
	l.write(LogLevelTrace, args...)
}

func (l stdoutLogger) Warn(args ...interface{}) {
	l.write(LogLevelWarn, args...)
}

func (l stdoutLogger) Error(args ...interface{}) {
	l.write(LogLevelError, args...)
}

func (l *stdoutLogger) Shutdown() error {
	return nil
}
