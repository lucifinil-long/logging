package logging

import (
	"bytes"
	"sync"
)

type loggerBuffer struct {
	bufferLock    sync.RWMutex
	bufferContent *bytes.Buffer
}

func (info *loggerBuffer) WriteString(str string) {
	info.bufferContent.WriteString(str)
}

func (info *loggerBuffer) WriteBuffer(bufferQueue chan loggerBuffer) {
	info.bufferLock.Lock()
	if info.bufferContent.Len() > 0 {
		bufferQueue <- *info
		info.bufferContent = bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
	}
	info.bufferLock.Unlock()
}

func newLoggerBuffer() *loggerBuffer {
	return &loggerBuffer{
		bufferContent: bytes.NewBuffer(make([]byte, 0, defaultBufferSize)),
	}
}
