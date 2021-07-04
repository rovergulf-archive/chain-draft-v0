package badgerdb

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type logger struct {
	zapLogger *zap.SugaredLogger
}

func NewLoggerFromZap() badger.Logger {
	lg, ok := viper.Get("logger").(*zap.SugaredLogger)
	if !ok {
		panic("no logger pointer saved in viper settings")
	}
	return &logger{
		zapLogger: lg,
	}
}

func (l logger) Errorf(msg string, args ...interface{}) {
	l.zapLogger.Errorf(msg, args...)
}
func (l logger) Warningf(msg string, args ...interface{}) {
	l.zapLogger.Warnf(msg, args...)
}
func (l logger) Infof(msg string, args ...interface{}) {
	l.zapLogger.Infof(msg, args...)
}
func (l logger) Debugf(msg string, args ...interface{}) {
	l.zapLogger.Debugf(msg, args...)
}
