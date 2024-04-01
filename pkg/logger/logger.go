package logger

// TODO would like to use slog when we can use go 1.21!

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
)

const (
	LevelNone = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelVerbose
	LevelDebug
)

var (
	DefaultLevel = 3
	logger       *RainbowLogger
)

type RainbowLogger struct {
	level    int
	Filename string
	handle   *os.File
}

// Start opens a file handle, if it's desired to write to file
func (l *RainbowLogger) Start() (*log.Logger, error) {
	f, err := os.OpenFile(l.Filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}
	logger := log.New(f, "", 0)
	l.handle = f
	return logger, nil
}

// Stop closes the file handle, if defined
func (l *RainbowLogger) Stop() error {
	if l.handle != nil {
		return l.handle.Close()
	}
	return nil
}

// Logging functions with formatting
func Infof(message ...any) error {
	return logger.logFormat(LevelInfo, message...)
}

func Errorf(message ...any) error {
	color.Set(color.FgRed)
	err := logger.logFormat(LevelError, message...)
	color.Unset()
	return err
}
func Debugf(message ...any) error {
	color.Set(color.FgBlue)
	err := logger.logFormat(LevelDebug, message...)
	color.Unset()
	return err

}
func Verbosef(message ...any) error {
	color.Set(color.FgMagenta)
	err := logger.logFormat(LevelVerbose, message...)
	color.Unset()
	return err

}
func Warningf(message ...any) error {
	color.Set(color.FgYellow)
	err := logger.logFormat(LevelWarning, message...)
	color.Unset()
	return err
}

// And without!
func Info(message string) error {
	return logger.log(LevelInfo, message)
}
func Error(message string) error {
	color.Set(color.FgRed)
	err := logger.log(LevelError, message)
	color.Unset()
	return err

}
func Debug(message string) error {
	color.Set(color.FgBlue)
	err := logger.log(LevelDebug, message)
	color.Unset()
	return err
}
func Verbose(message string) error {
	color.Set(color.FgMagenta)
	err := logger.log(LevelVerbose, message)
	color.Unset()
	return err
}
func Warning(message string) error {
	color.Set(color.FgYellow)
	err := logger.log(LevelWarning, message)
	color.Unset()
	return err
}

// log prints (without formatting) to the log
func (l *RainbowLogger) log(level int, message string) error {
	if l.Filename != "" {
		l.logToFile(level, message)
	}
	if level >= l.level {
		fmt.Println(message)
	}
	return nil
}

// logFormat is the shared class function for actually printing to the log
func (l *RainbowLogger) logFormat(level int, message ...any) error {
	if l.Filename != "" {
		l.logFormatToFile(level, message)
	}
	// Otherwise just print! Simple and dumb
	prolog := message[0].(string)
	rest := message[1:]
	if level <= l.level {
		fmt.Printf(prolog, rest...)
	}
	return nil
}

// logFormatToFile writes to file if the rainbow logger is set to do that
func (l *RainbowLogger) logFormatToFile(level int, message ...any) error {
	logger, err := l.Start()
	if err != nil {
		return err
	}
	// Assume the prolog (to be formatted) is at index 0
	prolog := message[0].(string)
	rest := message[1:]
	if level <= l.level {
		logger.Printf(prolog, rest...)
	}
	return l.Stop()
}

// logToFile writes to file if the rainbow logger is set to do that
func (l *RainbowLogger) logToFile(level int, message string) error {
	logger, err := l.Start()
	if err != nil {
		return err
	}
	if level <= l.level {
		logger.Println(message)
	}
	return l.Stop()
}

// WriteToFile will set the global level and filename
func WriteToFile(filename string) {
	logger.Filename = filename
}

// SetLevel sets the global logging level
func SetLevel(level int) {
	logger.level = level
}

func init() {
	logger = &RainbowLogger{level: LevelWarning}
}
