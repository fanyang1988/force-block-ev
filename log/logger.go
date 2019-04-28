package log

import (
	"go.uber.org/zap"
)

var logger = zap.NewNop()

func newLogger(production bool) (l *zap.Logger) {
	if production {
		l, _ = zap.NewProduction()
	} else {
		l, _ = zap.NewDevelopment()
	}
	return
}

// EnableLogging Enable logger for force block ev
func EnableLogging(production bool) {
	logger = newLogger(false)
}

// Logger get logger
func Logger() *zap.Logger {
	return logger
}

// SetLogger set logger
func SetLogger(l *zap.Logger) {
	logger = l
}
