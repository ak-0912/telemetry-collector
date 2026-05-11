// Package logging provides a configured standard logger for the collector.
package logging

import (
	"log"
	"os"
)

// New returns a logger that writes to stdout with timestamps and source location.
func New() *log.Logger {
	return log.New(os.Stdout, "telemetry-collector ", log.LstdFlags|log.Lshortfile)
}
