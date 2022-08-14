package logging

import "sync"

type synchronizedLogger struct {
	logger Logger
	mutex  sync.Mutex
}

func (sl *synchronizedLogger) Log(level LogLevel, text string, args ...interface{}) {
	sl.mutex.Lock()
	sl.logger.Log(level, text, args...)
	sl.mutex.Unlock()
}

func (sl *synchronizedLogger) MinLevel() LogLevel {
	return sl.logger.MinLevel()
}

func (sl *synchronizedLogger) IsConcurrent() bool {
	return true
}

func Synchronize(logger Logger) Logger {
	// If the logger is already thread-safe, return it as is.
	if logger.IsConcurrent() {
		return logger
	}

	return &synchronizedLogger{
		logger: logger,
	}
}
