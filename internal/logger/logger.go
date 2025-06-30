package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
)

func InitLogger() {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(zapcore.InfoLevel) // 设置默认日志级别

	var err error
	Logger, err = config.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}

func Sync() {
	_ = Logger.Sync()
}