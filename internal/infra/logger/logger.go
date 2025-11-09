package logger

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func Init() error {
	var err error
	Logger, err = zap.NewProduction()
	if err != nil {
		return err
	}
	return nil
}

func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}
