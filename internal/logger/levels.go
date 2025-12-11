package logger

type LogLevel int

const (
	LEVEL_DEBUG LogLevel = iota
	LEVEL_INFO
	LEVEL_WARN
	LEVEL_ERROR
	LEVEL_FATAL
)
