package logging

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type loggerInfo struct {
	filename      string
	buffer        *loggerBuffer
	bufferQueue   chan *cacheBuffer
	fsyncInterval time.Duration
	hour          time.Time
	fileOrder     int
	logFile       *os.File
	historyDir    string
	stopCh        chan bool
	endCh         chan bool
}

func newLoggerInfo(filename, level string) (*loggerInfo, error) {
	info := &loggerInfo{
		bufferQueue:   make(chan *cacheBuffer, 50000),
		fsyncInterval: time.Second,
		buffer:        newLoggerBuffer(),
		fileOrder:     0,
		historyDir:    "",
		stopCh:        make(chan bool),
		endCh:         make(chan bool),
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
		println(time.Now().Format("2006-01-02 15:04:05.000:"), "[NewLogger] openfile error : "+err.Error())
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
			println(time.Now().Format("2006-01-02 15:04:05.000:"), "[NeedSplit] FileSize: "+err.Error())
			if err = info.CreateFile(); err != nil {
				// if still failed, only output error
				println(time.Now().Format("2006-01-02 15:04:05.000:"), "[NeedSplit] CreateFile : "+err.Error())
			}
			return false, false
		}
		// only log error if error is not NoExistError
		println(time.Now().Format("2006-01-02 15:04:05.000:"), "[NeedSplit] FileSize: "+err.Error())
		return false, false
	}

	if size > maxFileSize {
		return true, false
	}

	return false, false
}

func (info *loggerInfo) Write(content string) {
	select {
	case <-info.stopCh:
		println(time.Now().Format("2006-01-02 15:04:05.000:"), "stopped logging yet, abandoned content:", content)
	default:
		info.buffer.WriteString(content)
	}
}

func (info *loggerInfo) WriteBufferToQueue() {
	// we don't add lock for only this go routine operates map
	ticker := time.NewTicker(info.fsyncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			info.buffer.WriteBuffer(info.bufferQueue, false)
		case <-info.stopCh:
			info.buffer.WriteBuffer(info.bufferQueue, true)
			return
		}
	}
}

func (info *loggerInfo) LoggerBackup(hour time.Time) {
	var oldFile string   //待备份文件
	var newFile string   //需要备份的新文件
	var backupDir string //备份的路径

	if info.historyDir == "" {
		return
	}
	backupDir = filepath.Join(info.historyDir, hour.Format(dateFormat))
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		os.MkdirAll(backupDir, 0777)
	}

	/* backup filename like saver-zwt-01-error.log.2014-09-10*/
	oldFile = info.filename + "." + hour.Format(hourFormat)
	if stat, err := os.Stat(oldFile); err == nil {
		newFile = filepath.Join(backupDir, stat.Name())
		if err := os.Rename(oldFile, newFile); err != nil {
			println(time.Now().Format("2006-01-02 15:04:05.000:"), "[LoggerBackup] os.Rename:"+err.Error())
		}
	}

	/* backup filename like saver-zwt-01-error.log.2014-09-10.{0/1...} */
	for i := 0; i < maxFileCount; i++ {
		oldFile = info.filename + "." + hour.Format(hourFormat) + "." + strconv.Itoa(i)
		if stat, err := os.Stat(oldFile); err == nil {
			newFile = filepath.Join(backupDir, stat.Name())
			if err := os.Rename(oldFile, newFile); err != nil {
				println(time.Now().Format("2006-01-02 15:04:05.000:"), "[LoggerBackup] os.Rename:"+err.Error())
			}
		}
	}
}

func (info *loggerInfo) split(isBackup bool) {
	info.logFile.Close()
	newFilename := info.filename + "." + info.hour.Format(hourFormat) + "." + strconv.Itoa(info.fileOrder%maxFileCount)
	_, fileErr := os.Stat(newFilename)
	if fileErr == nil {
		os.Remove(newFilename)
	}
	err := os.Rename(info.filename, newFilename)
	if err != nil {
		println(time.Now().Format("2006-01-02 15:04:05.000:"), "[FlushBufferQueue] Rename : "+err.Error())
	}
	if err = info.CreateFile(); err != nil {
		println(time.Now().Format("2006-01-02 15:04:05.000:"), "[FlushBufferQueue] CreateFile : "+err.Error())
	}

	info.fileOrder++
	if isBackup {
		info.fileOrder = 0
		go info.LoggerBackup(info.hour)
		info.hour, _ = time.Parse(hourFormat, time.Now().Format(hourFormat))
	}
}

func (info *loggerInfo) backup() {
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
		println(time.Now().Format("2006-01-02 15:04:05.000:"), "[FlushBufferQueue] Rename : "+err.Error())
	}
	if err = info.CreateFile(); err != nil {
		println(time.Now().Format("2006-01-02 15:04:05.000:"), "[FlushBufferQueue] CreateFile : "+err.Error())
	}

	info.fileOrder = 0
	go info.LoggerBackup(info.hour)
	info.hour, _ = time.Parse(hourFormat, time.Now().Format(hourFormat))
}

func (info *loggerInfo) FlushBufferQueue() {
	for {
		select {
		case cache := <-info.bufferQueue:
			// check whether need to split the log file
			isSplit, isBackup := info.NeedSplit()
			if isSplit {
				info.split(isBackup)
			} else {
				if isBackup {
					info.backup()
				}
			}

			// retry again if write failed
			if _, err := info.logFile.Write(cache.buffer.Bytes()); err != nil {
				println(time.Now().Format("2006-01-02 15:04:05.000:"), "[FlushBufferQueue] File.Write : "+err.Error())
				info.logFile.Write(cache.buffer.Bytes())
			}
			info.logFile.Sync()

			// check need to stop
			if cache.stop {
				println(time.Now().Format("2006-01-02 15:04:05.000:"), "received stop signal for", info.filename)
				info.logFile.Close()
				close(info.endCh)
				return
			}
		}
	}
}

func (info *loggerInfo) Close() {
	select {
	case <-info.stopCh:
	default:
		close(info.stopCh)
	}

	select {
	case <-info.endCh:
		println(time.Now().Format("2006-01-02 15:04:05.000:"), "stoped logging to", info.filename)
	}
}
