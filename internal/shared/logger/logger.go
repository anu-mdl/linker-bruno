package logger

import "log"

// Setup configures the default logger with standard flags
func Setup() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
