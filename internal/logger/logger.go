package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var currentLevel LogLevel = LEVEL_DEBUG
var logFile *os.File
var consoleLogger = log.New(os.Stdout, "", 0)

func Init(level LogLevel, filePath string) {
	currentLevel = level

	if filePath != "" {
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			logFile = f
		} else {
			consoleLogger.Println("Failed to open log file:", err)
		}
	}
}

func logMsg(level LogLevel, msg string) {
	if level < currentLevel {
		return
	}

	ts := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] [%s] %s", ts, levelName(level), decodeUnicode(msg))

	consoleLogger.Println(levelColor(level) + line + colorReset)

	if logFile != nil {
		fmt.Fprintln(logFile, line)
	}
}

func Debug(msg string) { logMsg(LEVEL_DEBUG, msg) }
func Info(msg string)  { logMsg(LEVEL_INFO, msg) }
func Warn(msg string)  { logMsg(LEVEL_WARN, msg) }
func Error(msg string) { logMsg(LEVEL_ERROR, msg) }
func Fatal(msg string) { logMsg(LEVEL_FATAL, msg); os.Exit(1) }
