package logging

import (
	"fmt"
	"testing"
	"time"
)

var (
	l Logger
)

func TestInit(t *testing.T) {
	logger, err := GetLogger("logs/logs", "logs/backuplogs", LogLevelError, "logging", "test")
	if err != nil {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05.999:"), "init logger failed with error", err)
		t.Fatal("init logger failed with error", err)
	}
	l = logger
}

func TestLogs(t *testing.T) {
	for i := 0; i < 10; i++ {
		l.Debug(fmt.Sprintf("this is #%d debug msg", i))
		l.Trace(fmt.Sprintf("this is #%d trace msg", i))
		l.Warn(fmt.Sprintf("this is #%d warn msg", i))
		l.Error(fmt.Sprintf("this is #%d error msg", i))
	}
}

func TestDown(t *testing.T) {
	time.Sleep(2 * time.Second)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.999:"), "unittests for logging are done.")
}
