package logging

import (
	"io"
	"log"
	"os"
)

var (
	defaultWriter    io.Writer
	additionalWriter io.Writer
)

type doubleLogger struct {
	defaultLogger    io.Writer
	additionalLogger io.Writer
}

// SetAdditionalOutput defines where we log to. This wraps the writer with another
// writer to allow a second logging location.
func SetAdditionalOutput(newAdditionalWriter io.Writer) {
	additionalWriter = newAdditionalWriter
	updateLogger()
}

// SetDefaultOutput defines the default location for logoutput. This can't
// be overriden by calling SetOutput.
func SetDefaultOutput(newDefaultWriter io.Writer) {
	defaultWriter = newDefaultWriter
	updateLogger()

}

func updateLogger() {
	if defaultWriter != nil && additionalWriter != nil {
		log.SetOutput(&doubleLogger{
			defaultLogger:    defaultWriter,
			additionalLogger: additionalWriter,
		})
	} else if defaultWriter != nil {
		log.SetOutput(defaultWriter)
	} else if additionalWriter != nil {
		log.SetOutput(additionalWriter)
	} else {
		log.SetOutput(os.Stdout)
	}
}

// Write redirects the output to both the default logger and the additional
// logger. If any is null, it is skipped. If any errors, an error is returned.
func (l *doubleLogger) Write(p []byte) (n int, err error) {
	var (
		count                int
		writeErrorDefault    error
		writeErrorAdditional error
	)
	if l.defaultLogger != nil {
		count, writeErrorDefault = l.defaultLogger.Write(p)
	}
	if l.additionalLogger != nil {
		count, writeErrorAdditional = l.additionalLogger.Write(p)
	}

	if writeErrorDefault != nil {
		return 0, writeErrorDefault
	}

	if writeErrorAdditional != nil {
		return 0, writeErrorAdditional
	}

	return count, nil
}
