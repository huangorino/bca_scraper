package utils

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is the global structured logger used across all files
var Logger = logrus.New()

// InitLogger sets up logrus with both file and console output, rotation, compression, timestamps
func InitLogger() {
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("[Error] Failed to create log directory: %v\n", err)
			os.Exit(1)
		}
	}

	logFilePath := fmt.Sprintf("%s/bca_crawler-%s.log", logDir, time.Now().Format("2006-01-02"))
	fileWriter := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    10,
		MaxBackups: 7,
		MaxAge:     30,
		Compress:   false,
	}

	multiWriter := io.MultiWriter(os.Stdout, fileWriter)
	Logger.SetOutput(multiWriter)

	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true,
		PadLevelText:    true,
	})

	Logger.SetLevel(logrus.InfoLevel)
	Logger.Info("----------------------------------------------------------")
	Logger.Infof("Log started at %s", time.Now().Format(time.RFC1123))
	Logger.Info("----------------------------------------------------------")
}
