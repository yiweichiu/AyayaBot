package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile *os.File
	mu      sync.Mutex
)

// Init initializes the logging system by creating the log directory
// and setting the output to a file named YYYYMMDD.log.
func Init() error {
	return Rotate()
}

// Rotate switches the log output to a new file based on the current date.
func Rotate() error {
	mu.Lock()
	defer mu.Unlock()

	// 1. Create log directory if it doesn't exist
	logDir := "log"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}
	}

	// 2. Prepare new log file name (YYYYMMDD.log)
	now := time.Now()
	fileName := fmt.Sprintf("%s.log", now.Format("20060102"))
	filePath := filepath.Join(logDir, fileName)

	// 3. Open new log file
	newFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// 4. Update log output
	oldFile := logFile
	log.SetOutput(newFile)
	logFile = newFile

	// 5. Close old file if it exists
	if oldFile != nil {
		oldFile.Close()
	}

	return nil
}

// Close closes the current log file handle.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
}
