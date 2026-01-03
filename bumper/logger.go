package bumper

import (
	"io"
	"log"
	"os"
)

var DebugLogger = log.New(io.Discard, "[DEBUG] ", log.LstdFlags)

func EnableDebugLogging() {
	DebugLogger.SetOutput(os.Stderr)
}

func DisableDebugLogging() {
	DebugLogger.SetOutput(io.Discard)
}
