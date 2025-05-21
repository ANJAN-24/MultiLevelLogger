package model

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (l LogLevel) IsAboveOrEqual(level LogLevel) bool {
	return l >= level
}

type LogPlatform interface {
	Log(level LogLevel, message string)
}

type ConsolePlatform struct{}

func (p *ConsolePlatform) Log(level LogLevel, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] [%s] %s\n", timestamp, level.String(), message)
}

type TextFilePlatform struct {
	FilePath string
	Mu       sync.Mutex
}

func NewTextFilePlatform(filePath string) *TextFilePlatform {
	return &TextFilePlatform{
		FilePath: filePath,
	}
}

func (p *TextFilePlatform) Log(level LogLevel, message string) {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level.String(), message)

	file, err := os.OpenFile(p.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(logEntry); err != nil {
		fmt.Printf("Failed to write to log file: %v\n", err)
	}
}

type NetworkPlatform struct {
	Logs []string
	Mu   sync.Mutex
}

func NewNetworkPlatform() *NetworkPlatform {
	return &NetworkPlatform{
		Logs: make([]string, 0),
	}
}

func (p *NetworkPlatform) Log(level LogLevel, message string) {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), message)
	p.Logs = append(p.Logs, logEntry)
}

func (p *NetworkPlatform) GetLogs() []string {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	result := make([]string, len(p.Logs))
	copy(result, p.Logs)
	return result
}

type AsyncTask struct {
	Platform LogPlatform
	Level    LogLevel
	Message  string
}

type Config struct {
	MinLogLevel string `json:"min_log_level"`
	LogFilePath string `json:"log_file_path"`
	UseTextFile bool   `json:"use_text_file"`
	UseNetwork  bool   `json:"use_network"`
}

func (c *Config) GetLogLevel() LogLevel {
	switch c.MinLogLevel {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}
