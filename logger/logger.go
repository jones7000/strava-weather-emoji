package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	Logger    *log.Logger
	logFile   *os.File
	logTarget string
)

func InitLogger(target, filePath string) error {
	logTarget = target

	// Ermitteln des Verzeichnisses des Hauptprogramms (main.go)
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %v", err)
	}
	dir := filepath.Dir(exePath)

	if filePath == "" {
		filePath = filepath.Join(dir, "app.log")
	} else {
		filePath = filepath.Join(dir, filePath)
	}

	if target == "file" {
		var err error
		logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening log file: %v", err)
		}
		Logger = log.New(logFile, "", log.LstdFlags)
	} else {
		Logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	return nil
}

// LogMessage loggt eine formatierte Nachricht
func LogMessage(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	switch logTarget {
	case "console":
		fmt.Println(message)
	case "file":
		Logger.Println(message)
	default:
		Logger.Println(message)
	}
}

// CloseLogger schließt die Logdatei (falls geöffnet)
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
