package main

import (
	"fmt"
	"logPlatform/logger"
	"logPlatform/logger/model"
	"time"
)

func main() {
	config, err := logger.LoadConfig("config.json")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		config = &model.Config{
			MinLogLevel: "DEBUG",
			LogFilePath: "logs.txt",
			UseTextFile: true,
			UseNetwork:  true,
		}
	}

	log := logger.CreateLoggerFromConfig(config)
	defer log.Shutdown()

	fmt.Println("--- Initial Log Level: DEBUG ---")
	log.Debug("This is a debug message with value: %d", 42)
	log.Info("This is my first log with Id: %s", "ABC123")
	log.Warning("This is a warning message: %s", "Something might be wrong")
	log.Error("I got the error: %s", "Connection failed")

	printNetworkLogs(log)

	fmt.Println("\n--- Changing log level to INFO ---")
	log.SetLevel(model.INFO)

	log.Info("This info will be logged")
	log.Warning("This warning will be logged")
	log.Error("This error will be logged")

	printNetworkLogs(log)

	fmt.Println("\n--- Changing log level to WARNING ---")
	log.SetLevel(model.WARNING)

	log.Warning("This warning will be logged")
	log.Error("This error will be logged")

	printNetworkLogs(log)

	fmt.Println("\n--- Changing log level to ERROR ---")
	log.SetLevel(model.ERROR)

	log.Error("Only errors will be logged")

	printNetworkLogs(log)

	fmt.Println("\n--- Manually detaching network platform ---")
	log.DetachNetworkPlatform()
	log.Error("This error won't go to network")

	fmt.Println("\n--- Manually attaching network platform ---")
	log.AttachNetworkPlatform()
	log.Error("This error will go to network")

	time.Sleep(100 * time.Millisecond)
	log.WaitForAsyncLogs()
	printNetworkLogs(log)
}

func printNetworkLogs(log *logger.Logger) {
	log.WaitForAsyncLogs()
	networkLogs := log.GetNetworkLogs()
	if networkLogs != nil {
		fmt.Println("\n--- Network Logs ---")
		for i, entry := range networkLogs {
			fmt.Printf("%d: %s\n", i+1, entry)
		}
	} else {
		fmt.Println("\n--- No Network Logs Available ---")
	}
}
