package logging

import (
	"log"
	"os"
)

func New() *log.Logger {
	return log.New(os.Stdout, "telemetry-collector ", log.LstdFlags|log.Lshortfile)
}
