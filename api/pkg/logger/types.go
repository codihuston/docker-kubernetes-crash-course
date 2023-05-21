package logger

// ALL < TRACE < DEBUG < INFO < WARN < ERROR < FATAL < OFF.
type Logger interface {
	Config()
	Trace(...any)
	Debug(...any)
	Info(...any)
	Warn(...any)
	Error(...any)
	Fatal(...any)
}
