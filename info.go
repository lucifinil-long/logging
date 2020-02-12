package logging

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type loggerInfo struct {
	filename       string
	bufferInfoLock sync.RWMutex
	buffer         *loggerBuffer
	bufferQueue    chan loggerBuffer
	fsyncInterval  time.Duration
	hour           time.Time
	fileOrder      int
	logFile        *os.File
	backupDir      string
}

func newLoggerInfo(filename, level string) (*loggerInfo, error) {
	info := &loggerInfo{
		bufferQueue:   make(chan loggerBuffer, 50000),
		fsyncInterval: time.Second,
		buffer:        newLoggerBuffer(),
		fileOrder:     0,
		backupDir:     "",
	}

	t, _ := time.Parse(hourFormat, time.Now().Format(hourFormat))
	info.hour = t

	// 直接调用write写日志的文件名，用原始的文件名
	if len(level) == 0 {
		info.filename = filename
	} else {
		info.filename = filename + "-" + level + ".log"
	}

	err := info.CreateFile()
	if err != nil {
		println("[NewLogger] openfile error : " + err.Error())
		return nil, err
	}
	return info, nil
}

func (info *loggerInfo) FileSize() (int64, error) {
	f, err := os.Stat(info.filename)
	if err != nil {
		return 0, err
	}
	return f.Size(), nil
}

func (info *loggerInfo) CreateFile() error {
	var err error
	info.logFile, err = os.OpenFile(info.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	return err
}

func (info *loggerInfo) NeedSplit() (split bool, backup bool) {
	t, _ := time.Parse(hourFormat, time.Now().Format(hourFormat))
	if t.After(info.hour) {
		return false, true
	}

	size, err := info.FileSize()
	if err != nil {
		if os.IsNotExist(err) {
			// file is not exist, recreates file
			println("[NeedSplit] FileSize: " + err.Error())
			if err = info.CreateFile(); err != nil {
				// if still failed, only output error
				println("[NeedSplit] CreateFile : " + err.Error())
			}
			return false, false
		}
		// only log error if error is not NoExistError
		println("[NeedSplit] FileSize: " + err.Error())
		return false, false
	}

	if size > maxFileSize {
		return true, false
	}

	return false, false
}

func (info *loggerInfo) Write(content string) {
	info.bufferInfoLock.Lock()
	info.buffer.WriteString(content)
	info.bufferInfoLock.Unlock()
}

func (info *loggerInfo) WriteBufferToQueue() {
	// we don't add lock for only this go routine operates map
	ticker := time.NewTicker(info.fsyncInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		info.bufferInfoLock.RLock()
		info.buffer.WriteBuffer(info.bufferQueue)
		info.bufferInfoLock.RUnlock()
	}
}

func (info *loggerInfo) LoggerBackup(hour time.Time) {
	var oldFile string   //待备份文件
	var newFile string   //需要备份的新文件
	var backupDir string //备份的路径

	if info.backupDir == "" {
		return
	}
	backupDir = filepath.Join(info.backupDir, hour.Format(dateFormat))
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		os.MkdirAll(backupDir, 0777)
	}

	/* backup filename like saver-zwt-01-error.log.2014-09-10*/
	oldFile = info.filename + "." + hour.Format(hourFormat)
	if stat, err := os.Stat(oldFile); err == nil {
		newFile = filepath.Join(backupDir, stat.Name())
		if err := os.Rename(oldFile, newFile); err != nil {
			println("[LoggerBackup] os.Rename:" + err.Error())
		}
	}

	/* backup filename like saver-zwt-01-error.log.2014-09-10.{0/1...} */
	for i := 0; i < maxFileCount; i++ {
		oldFile = info.filename + "." + hour.Format(hourFormat) + "." + strconv.Itoa(i)
		if stat, err := os.Stat(oldFile); err == nil {
			newFile = filepath.Join(backupDir, stat.Name())
			if err := os.Rename(oldFile, newFile); err != nil {
				println("[LoggerBackup] os.Rename:" + err.Error())
			}
		}
	}
}

func (info *loggerInfo) FlushBufferQueue() {
	for {
		select {
		case buffer := <-info.bufferQueue:
			// check whether need to split the log file
			isSplit, isBackup := info.NeedSplit()
			if isSplit {
				info.logFile.Close()
				newFilename := info.filename + "." + info.hour.Format(hourFormat) + "." + strconv.Itoa(info.fileOrder%maxFileCount)
				_, fileErr := os.Stat(newFilename)
				if fileErr == nil {
					os.Remove(newFilename)
				}
				err := os.Rename(info.filename, newFilename)
				if err != nil {
					println("[FlushBufferQueue] Rename : " + err.Error())
				}
				if err = info.CreateFile(); err != nil {
					println("[FlushBufferQueue] CreateFile : " + err.Error())
				}

				info.fileOrder++
				if isBackup {
					info.fileOrder = 0
					go info.LoggerBackup(info.hour)
					info.hour, _ = time.Parse(hourFormat, time.Now().Format(hourFormat))
				}
			} else {
				if isBackup {
					info.logFile.Close()

					var newFilename string
					if info.fileOrder == 0 {
						newFilename = info.filename + "." + info.hour.Format(hourFormat)
					} else {
						newFilename = info.filename + "." + info.hour.Format(hourFormat) + "." + strconv.Itoa(info.fileOrder%maxFileCount)
					}

					_, fileErr := os.Stat(newFilename)
					if fileErr == nil {
						os.Remove(newFilename)
					}
					err := os.Rename(info.filename, newFilename)
					if err != nil {
						println("[FlushBufferQueue] Rename : " + err.Error())
					}
					if err = info.CreateFile(); err != nil {
						println("[FlushBufferQueue] CreateFile : " + err.Error())
					}

					info.fileOrder = 0
					go info.LoggerBackup(info.hour)
					info.hour, _ = time.Parse(hourFormat, time.Now().Format(hourFormat))
				}
			}

			// retry again if write failed
			if _, err := info.logFile.Write(buffer.bufferContent.Bytes()); err != nil {
				println("[FlushBufferQueue] File.Write : " + err.Error())
				info.logFile.Write(buffer.bufferContent.Bytes())
			}
			info.logFile.Sync()
		}
	}
}
