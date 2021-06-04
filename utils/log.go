package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var defaultSugarLogger *zap.SugaredLogger

func initLog() {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	logger := zap.New(core, zap.AddCaller())
	defaultSugarLogger = logger.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./login.log",
		MaxSize:    500000,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}

	return zapcore.AddSync(lumberJackLogger)
}

func Logger() *zap.SugaredLogger {
	if defaultSugarLogger == nil {
		initLog()
	}

	return defaultSugarLogger
}
