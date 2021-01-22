package logging

import (
	"bytes"
	"sync"
)

type loggerBuffer struct {
	bufferLock    sync.RWMutex
	bufferContent *bytes.Buffer
}

type cacheBuffer struct {
	buffer *bytes.Buffer
	stop   bool
}

func (info *loggerBuffer) WriteString(str string) {
	info.bufferLock.Lock()
	info.bufferContent.WriteString(str)
	info.bufferLock.Unlock()
}

func (info *loggerBuffer) WriteBuffer(bufferQueue chan *cacheBuffer, stop bool) {
	info.bufferLock.Lock()
	if info.bufferContent.Len() > 0 || stop {
		bufferQueue <- &cacheBuffer{
			buffer: info.bufferContent,
			stop:   stop,
		}
		info.bufferContent = bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
	}
	info.bufferLock.Unlock()
}

func newLoggerBuffer() *loggerBuffer {
	return &loggerBuffer{
		bufferContent: bytes.NewBuffer(make([]byte, 0, defaultBufferSize)),
	}
}
