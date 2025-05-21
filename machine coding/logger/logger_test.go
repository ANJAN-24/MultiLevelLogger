package logger

import (
	"logPlatform/logger/model"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLogLevels(t *testing.T) {
	testFilePath := "test_logs.txt"

	os.Remove(testFilePath)

	logger := NewLogger(model.DEBUG, testFilePath)
	defer logger.Shutdown()
	defer os.Remove(testFilePath)

	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warning("Warning message")
	logger.Error("Error message")

	logger.WaitForAsyncLogs()

	fileContent, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	content := string(fileContent)

	if !strings.Contains(content, "Info message") {
		t.Error("Log file should contain INFO message")
	}
	if !strings.Contains(content, "Warning message") {
		t.Error("Log file should contain WARNING message")
	}
	if !strings.Contains(content, "Error message") {
		t.Error("Log file should contain ERROR message")
	}
	if strings.Contains(content, "Debug message") {
		t.Error("Log file should NOT contain DEBUG message")
	}
}

func TestNetworkLogLevels(t *testing.T) {
	logger := NewLogger(model.DEBUG, "")
	defer logger.Shutdown()

	logger.Debug("Debug message for network test")
	logger.Info("Info message for network test")
	logger.Warning("Warning message for network test")
	logger.Error("Error message for network test")

	logger.WaitForAsyncLogs()

	logs := logger.GetNetworkLogs()

	for _, logEntry := range logs {
		if strings.Contains(logEntry, "Debug message") {
			t.Error("Network logs should NOT contain DEBUG messages")
		}
		if strings.Contains(logEntry, "Info message") {
			t.Error("Network logs should NOT contain INFO messages")
		}
	}

	warningCount := 0
	errorCount := 0

	for _, log := range logs {
		if strings.Contains(log, "Warning message") {
			warningCount++
		}
		if strings.Contains(log, "Error message") {
			errorCount++
		}
	}

	if warningCount != 1 {
		t.Errorf("Expected 1 WARNING log in network, got %d", warningCount)
	}
	if errorCount != 1 {
		t.Errorf("Expected 1 ERROR log in network, got %d", errorCount)
	}

	if len(logs) != 2 {
		t.Errorf("Expected exactly 2 logs in network (WARNING + ERROR), got %d", len(logs))
	}
}

func TestChangingLogLevel(t *testing.T) {
	logger := NewLogger(model.DEBUG, "")
	defer logger.Shutdown()


	logger.Debug("Debug 1")
	logger.Info("Info 1")
	logger.Warning("Warning 1")
	logger.Error("Error 1")

	logger.SetLevel(model.WARNING)

	logger.Debug("Debug 2")
	logger.Info("Info 2")
	logger.Warning("Warning 2")
	logger.Error("Error 2")

	logger.WaitForAsyncLogs()

	logs := logger.GetNetworkLogs()

	warningCount := 0
	errorCount := 0
	debugCount := 0
	infoCount := 0

	for _, log := range logs {
		if strings.Contains(log, "Debug") {
			debugCount++
		}
		if strings.Contains(log, "Info") {
			infoCount++
		}
		if strings.Contains(log, "Warning") {
			warningCount++
		}
		if strings.Contains(log, "Error") {
			errorCount++
		}
	}

	if debugCount != 0 {
		t.Errorf("Expected 0 DEBUG logs in network, got %d", debugCount)
	}
	if infoCount != 0 {
		t.Errorf("Expected 0 INFO logs in network, got %d", infoCount)
	}
	if warningCount != 2 {
		t.Errorf("Expected 2 WARNING logs in network, got %d", warningCount)
	}
	if errorCount != 2 {
		t.Errorf("Expected 2 ERROR logs in network, got %d", errorCount)
	}
}

func TestAsyncLogging(t *testing.T) {
	logger := NewLogger(model.DEBUG, "")

	for i := 0; i < 10; i++ {
		logger.Warning("Test message")
	}

	initialLogs := logger.GetNetworkLogs()

	time.Sleep(100 * time.Millisecond)
	logger.WaitForAsyncLogs()

	finalLogs := logger.GetNetworkLogs()

	if len(finalLogs) != 10 {
		t.Errorf("Expected 10 logs after waiting, got %d", len(finalLogs))
	}

	if len(initialLogs) > 0 && len(initialLogs) < 10 {
		t.Logf("Confirmed async behavior: initially had %d logs, later had %d",
			len(initialLogs), len(finalLogs))
	}

	logger.Shutdown()
}

func TestPlatformAttachDetach(t *testing.T) {
	logger := NewLogger(model.ERROR, "")
	defer logger.Shutdown()

	if logger.GetNetworkLogs() != nil {
		t.Error("Network logs should be nil for ERROR level logger")
	}

	logger.AttachNetworkPlatform()
	logger.Error("Error after attach")
	logger.WaitForAsyncLogs()

	logs1 := logger.GetNetworkLogs()
	if len(logs1) != 1 {
		t.Errorf("Expected 1 log after attaching, got %d", len(logs1))
	}

	logger.DetachNetworkPlatform()

	if logger.GetNetworkLogs() != nil {
		t.Error("Network logs should be nil after detaching platform")
	}

	logger.AttachNetworkPlatform()
	logger.Error("Error after reattach")
	logger.WaitForAsyncLogs()

	logs2 := logger.GetNetworkLogs()
	if len(logs2) != 1 {
		t.Errorf("Expected 1 log after reattaching, got %d", len(logs2))
	}
}
