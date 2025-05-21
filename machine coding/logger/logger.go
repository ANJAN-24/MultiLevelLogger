package logger

import (
	"fmt"
	"logPlatform/logger/model"
	"os"
	"sync"
)

type Logger struct {
	minLevel         model.LogLevel
	consolePlatform  *model.ConsolePlatform
	textFilePlatform *model.TextFilePlatform
	networkPlatform  *model.NetworkPlatform
	asyncQueue       chan model.AsyncTask
	asyncWaitGroup   sync.WaitGroup
	mu               sync.Mutex
}

func NewLogger(minLevel model.LogLevel, logFilePath string) *Logger {
	logger := &Logger{
		minLevel:        minLevel,
		consolePlatform: &model.ConsolePlatform{},
		asyncQueue:      make(chan model.AsyncTask, 100),
	}

	go logger.processAsyncTasks()

	if logFilePath != "" && minLevel <= model.INFO {
		clearLogFile(logFilePath)
		logger.textFilePlatform = model.NewTextFilePlatform(logFilePath)
	}

	if minLevel <= model.WARNING {
		logger.networkPlatform = model.NewNetworkPlatform()
	}

	return logger
}

func clearLogFile(filePath string) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err == nil {
		file.Close()
	}
}

func (l *Logger) processAsyncTasks() {
	for task := range l.asyncQueue {
		task.Platform.Log(task.Level, task.Message)
		l.asyncWaitGroup.Done()
	}
}

func (l *Logger) SetLevel(level model.LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level

	if level <= model.INFO && l.textFilePlatform == nil && l.GetLogFilePath() != "" {
		l.textFilePlatform = model.NewTextFilePlatform(l.GetLogFilePath())
	} else if level > model.INFO && l.textFilePlatform != nil {
		l.textFilePlatform = nil
	}

	if level <= model.WARNING && l.networkPlatform == nil {
		l.networkPlatform = model.NewNetworkPlatform()
	} else if level > model.WARNING && l.networkPlatform != nil {
		l.networkPlatform = nil
	}
}

func (l *Logger) GetLogFilePath() string {
	if l.textFilePlatform == nil {
		return ""
	}
	return l.textFilePlatform.FilePath
}

func (l *Logger) AttachNetworkPlatform() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.networkPlatform == nil {
		l.networkPlatform = model.NewNetworkPlatform()
	}
}

func (l *Logger) DetachNetworkPlatform() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.networkPlatform = nil
}

func (l *Logger) AttachTextFilePlatform(filePath string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	clearLogFile(filePath)
	l.textFilePlatform = model.NewTextFilePlatform(filePath)
}

func (l *Logger) DetachTextFilePlatform() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.textFilePlatform = nil
}

func (l *Logger) GetNetworkLogs() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.networkPlatform == nil {
		return nil
	}
	return l.networkPlatform.GetLogs()
}

func (l *Logger) logWithPlatform(platform model.LogPlatform, level model.LogLevel, message string, async bool) {
	if async && platform == l.networkPlatform {
		l.asyncWaitGroup.Add(1)
		l.asyncQueue <- model.AsyncTask{
			Platform: platform,
			Level:    level,
			Message:  message,
		}
	} else {
		platform.Log(level, message)
	}
}

func (l *Logger) log(level model.LogLevel, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !level.IsAboveOrEqual(l.minLevel) {
		return
	}

	l.logWithPlatform(l.consolePlatform, level, message, false)

	if level.IsAboveOrEqual(model.INFO) && l.textFilePlatform != nil {
		l.logWithPlatform(l.textFilePlatform, level, message, false)
	}

	if level.IsAboveOrEqual(model.WARNING) && l.networkPlatform != nil {
		l.logWithPlatform(l.networkPlatform, level, message, true)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(model.DEBUG, message)
}

func (l *Logger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(model.INFO, message)
}

func (l *Logger) Warning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(model.WARNING, message)
}

func (l *Logger) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(model.ERROR, message)
}

func (l *Logger) WaitForAsyncLogs() {
	l.asyncWaitGroup.Wait()
}

func (l *Logger) Shutdown() {
	l.WaitForAsyncLogs()
	close(l.asyncQueue)
}
