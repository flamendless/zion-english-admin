package logs

import (
	"errors"
	"fmt"
	"sync"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var loggerOnce sync.Once
var logger *zap.Logger

func InitLog() {
	var config zap.Config

	config = zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.Level(0))
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	newLogger, err := config.Build()
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		err := newLogger.Sync()
		if err != nil && !errors.Is(err, syscall.ENOTTY) {
			fmt.Println(err)
		}
	}()

	logger = newLogger
}

func Log() *zap.Logger {
	loggerOnce.Do(func() {
		InitLog()
	})
	return logger
}
