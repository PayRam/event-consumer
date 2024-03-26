package logger

import (
	"log"
	"os"
)

func init() {
	// Initialize logger with default settings.
	InitLogger()
}

// Initialize the logger with default settings.
// This function should be called at the beginning of your application.
func InitLogger() {
	// Set the output destination of the logger
	log.SetOutput(os.Stdout) // You can change this to a file or any io.Writer

	// Set a prefix that will be used for all log messages
	log.SetPrefix("LOG: ")

	// Define what to include in the log output
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
}

// Info logs a message at the Info level.
func Info(format string, v ...interface{}) {
	log.Printf("INFO: "+format, v...)
}

// Error logs a message at the Error level.
func Error(format string, v ...interface{}) {
	log.Printf("ERROR: "+format, v...)
}

// Debug logs a message at the Debug level.
// Note: In a production setting, you might want to disable debug logs or control them via an environment variable.
func Debug(format string, v ...interface{}) {
	log.Printf("DEBUG: "+format, v...)
}

// Warn logs a message at the Warn level.
func Warn(format string, v ...interface{}) {
	log.Printf("WARN: "+format, v...)
}
