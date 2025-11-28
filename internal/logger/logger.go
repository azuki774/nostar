package logger

import (
	"go.uber.org/zap"
)

func Load() *zap.Logger {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
	return logger
}
