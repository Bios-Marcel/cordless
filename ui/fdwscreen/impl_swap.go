// +build linux darwin netbsd openbsd

package fdwscreen

import (
	"os"

	"github.com/MOZGIII/fdswap-go"
)

type fdSwapper struct {
	logFilePath string

	logFile *os.File

	swappedFdHandleStdout *fdswap.SwappedFdHandle
	swappedFdHandleStderr *fdswap.SwappedFdHandle
}

func newFdSwapper() (*fdSwapper, error) {
	logFilePath := os.Getenv("CORDLESS_LOG_FILE_PATH")
	if logFilePath == "" {
		// TODO: use more sensible default.
		logFilePath = "log.txt"
	}

	return &fdSwapper{
		logFilePath: logFilePath,

		logFile: nil,

		swappedFdHandleStdout: nil,
		swappedFdHandleStderr: nil,
	}, nil
}

func (s *fdSwapper) InitSwap() {
	s.clear()

	var err error

	s.logFile, err = os.OpenFile(s.logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	s.swappedFdHandleStdout, err = fdswap.SwapFiles(os.Stdout, s.logFile)
	if err != nil {
		panic(err)
	}

	s.swappedFdHandleStderr, err = fdswap.SwapFiles(os.Stderr, s.logFile)
	if err != nil {
		panic(err)
	}
}

func (s *fdSwapper) FiniSwap() {
	s.clear()
}

func (s *fdSwapper) clear() {
	if s.swappedFdHandleStdout != nil {
		if err := s.swappedFdHandleStdout.Restore(); err != nil {
			panic(err)
		}
		s.swappedFdHandleStdout = nil
	}

	if s.swappedFdHandleStderr != nil {
		if err := s.swappedFdHandleStderr.Restore(); err != nil {
			panic(err)
		}
		s.swappedFdHandleStderr = nil
	}

	if s.logFile != nil {
		if err := s.logFile.Close(); err != nil {
			panic(err)
		}
		s.logFile = nil
	}
}
