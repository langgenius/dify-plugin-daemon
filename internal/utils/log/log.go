package log

/*
	log module is used to write log info to log file
	open a log file when log was created, and close it when log was destroyed
*/

import (
	"fmt"
	go_log "log"
	"os"
)

var show_log bool = true
var logger = go_log.New(os.Stdout, "", go_log.Ldate|go_log.Ltime|go_log.Lshortfile)

const (
	LOG_LEVEL_DEBUG_COLOR = "\033[34m"
	LOG_LEVEL_INFO_COLOR  = "\033[32m"
	LOG_LEVEL_WARN_COLOR  = "\033[33m"
	LOG_LEVEL_ERROR_COLOR = "\033[31m"
	LOG_LEVEL_COLOR_END   = "\033[0m"
)

func writeLog(level string, format string, stdout bool, v ...interface{}) {
	//write log
	format = fmt.Sprintf("["+level+"]"+format, v...)

	if show_log && stdout {
		if level == "DEBUG" {
			logger.Output(3, LOG_LEVEL_DEBUG_COLOR+format+LOG_LEVEL_COLOR_END)
		} else if level == "INFO" {
			logger.Output(3, LOG_LEVEL_INFO_COLOR+format+LOG_LEVEL_COLOR_END)
		} else if level == "WARN" {
			logger.Output(3, LOG_LEVEL_WARN_COLOR+format+LOG_LEVEL_COLOR_END)
		} else if level == "ERROR" {
			logger.Output(3, LOG_LEVEL_ERROR_COLOR+format+LOG_LEVEL_COLOR_END)
		} else if level == "PANIC" {
			logger.Output(3, LOG_LEVEL_ERROR_COLOR+format+LOG_LEVEL_COLOR_END)
			panic(format)
		}
	}
}

func SetShowLog(show bool) {
	show_log = show
}

func Debug(format string, v ...interface{}) {
	writeLog("DEBUG", format, true, v...)
}

func Info(format string, v ...interface{}) {
	writeLog("INFO", format, true, v...)
}

func Warn(format string, v ...interface{}) {
	writeLog("WARN", format, true, v...)
}

func Error(format string, v ...interface{}) {
	writeLog("ERROR", format, true, v...)
}

func Panic(format string, v ...interface{}) {
	writeLog("PANIC", format, true, v...)
}
