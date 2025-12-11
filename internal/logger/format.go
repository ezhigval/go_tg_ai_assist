package logger

var (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGreen  = "\033[32m"
	colorPurple = "\033[35m"
)

func levelColor(level LogLevel) string {
	switch level {
	case LEVEL_DEBUG:
		return colorBlue
	case LEVEL_INFO:
		return colorGreen
	case LEVEL_WARN:
		return colorYellow
	case LEVEL_ERROR:
		return colorRed
	case LEVEL_FATAL:
		return colorPurple
	}
	return colorReset
}

func levelName(level LogLevel) string {
	switch level {
	case LEVEL_DEBUG:
		return "DEBUG"
	case LEVEL_INFO:
		return "INFO"
	case LEVEL_WARN:
		return "WARN"
	case LEVEL_ERROR:
		return "ERROR"
	case LEVEL_FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}
