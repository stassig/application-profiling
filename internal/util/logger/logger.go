package logger

import (
	"log"
)

// Info logs an informational message
func Info(message string) {
	log.Printf("[INFO] %s\n", message)
}

// Warning logs a warning message
func Warning(message string) {
	log.Printf("[WARNING] %s\n", message)
}

// Debug logs a debug message
func Debug(message string) {
	log.Printf("[DEBUG] %s\n", message)
}

// Error logs a critical error and exits the program
func Error(err error, message string) {
	if err != nil {
		log.Fatalf("[ERROR] %s: %v\n", message, err)
	}
}