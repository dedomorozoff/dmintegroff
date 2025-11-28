package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func Init() {
	// Set log level
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	Log.SetLevel(logLevel)

	// Set output
	logFile := os.Getenv("LOG_FILE")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			Log.SetOutput(io.MultiWriter(os.Stdout, file))
		} else {
			Log.Info("Failed to log to file, using default stderr")
		}
	} else {
		Log.SetOutput(os.Stdout)
	}

	// Set formatter
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}
