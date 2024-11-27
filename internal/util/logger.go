package util

import (
	"log"
)

// LogError checks for an error and logs it if present
func LogError(err error, message string) {
	if err != nil {
		log.Fatalf("[ERROR] %s: %v\n", message, err)
	}
}
