package logger

import "go.uber.org/zap"

var Log *zap.Logger

// InitLogger 初始化 Zap 日志
func InitLogger() *zap.Logger {
	Logger, _ := zap.NewProduction()
	Log = Logger
	return Logger
}
