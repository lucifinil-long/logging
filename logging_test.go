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
	config := &Config{
		LogDir:     DefaultLogDir,
		HistoryDir: DefaultHistoryDir,
		Level:      LogLevelDebug,
	}

	logger, err := GetLogger("logging", "test", config)
	if err != nil {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05.999:"), "init logger failed with error", err)
		t.Fatal("init logger failed with error", err)
	}
	l = logger
}

func TestLogs(_ *testing.T) {
	for i := 0; i < 10; i++ {
		l.Debug(fmt.Sprintf("this is #%d debug msg", i))
		l.Trace(fmt.Sprintf("this is #%d trace msg", i))
		l.Warn(fmt.Sprintf("this is #%d warn msg", i))
		l.Error(fmt.Sprintf("this is #%d error msg", i))
	}
}

func TestDown(_ *testing.T) {
	err := l.Shutdown()
	l.Debug("log request after shutdown")
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.999:"), "unittests for logging are done. error: ", err)
}
