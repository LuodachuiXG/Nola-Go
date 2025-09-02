package logger

import "go.uber.org/zap"

var Logger *zap.Logger

// InitLogger 初始化 Zap 日志
func InitLogger() *zap.Logger {
	Logger, _ := zap.NewProduction()
	return Logger
}
