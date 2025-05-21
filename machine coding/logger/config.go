package logger

import (
	"encoding/json"
	"logPlatform/logger/model"
	"os"
)

func LoadConfig(configPath string) (*model.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config model.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func CreateLoggerFromConfig(config *model.Config) *Logger {
	if config.LogFilePath != "" {
		clearLogFile(config.LogFilePath)
	}

	logger := NewLogger(config.GetLogLevel(), config.LogFilePath)

	if config.UseNetwork {
		logger.AttachNetworkPlatform()
	}

	return logger
}
