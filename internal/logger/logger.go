package logger

import (
	"fmt" // Required for Sprintf and Sprintln
	"log"
	"os"
)

var (
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	debugMode   bool
)

// Init 初始化日志记录器
func Init(enableDebug bool) {
	debugMode = enableDebug

	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	if debugMode {
		debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		// If debug mode is not enabled, discard debug logs
		debugLogger = log.New(discardWriter{}, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

// discardWriter implements the io.Writer interface but discards all written data.
type discardWriter struct{}

func (discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Info logs a message at InfoLevel
func Info(v ...interface{}) {
	infoLogger.Output(3, fmt.Sprintln(v...))
}

// Infof logs a formatted message at InfoLevel
func Infof(format string, v ...interface{}) {
	infoLogger.Output(3, fmt.Sprintf(format, v...))
}

// Warn 记录警告级别日志
func Warn(v ...interface{}) {
	warnLogger.Output(3, fmt.Sprintln(v...))
}

// Warnf 格式化记录警告级别日志
func Warnf(format string, v ...interface{}) {
	warnLogger.Output(3, fmt.Sprintf(format, v...))
}

// Error logs a message at the error level.
func Error(v ...interface{}) {
	errorLogger.Output(3, fmt.Sprintln(v...))
}

// Errorf logs a formatted message at the error level.
func Errorf(format string, v ...interface{}) {
	errorLogger.Output(3, fmt.Sprintf(format, v...))
}

// Fatal logs a message at the fatal level and exits the program.
func Fatal(v ...interface{}) {
	errorLogger.Output(3, fmt.Sprintln(v...))
	os.Exit(1)
}

// Fatalf logs a formatted message at the fatal level and exits the program.
func Fatalf(format string, v ...interface{}) {
	errorLogger.Output(3, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Debug 记录调试级别日志 (仅在 debugMode 为 true 时)
func Debug(v ...interface{}) {
	// No need to check debugMode here, as debugLogger is already a discard writer if not enabled
	debugLogger.Output(3, fmt.Sprintln(v...))
}

// Debugf 格式化记录调试级别日志 (仅在 debugMode 为 true 时)
func Debugf(format string, v ...interface{}) {
	// No need to check debugMode here, as debugLogger is already a discard writer if not enabled
	debugLogger.Output(3, fmt.Sprintf(format, v...))
}
