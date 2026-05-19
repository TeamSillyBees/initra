package asynqadapter

import "go.uber.org/zap"

type zapLogger struct {
	logger *zap.SugaredLogger
}

func newZapLogger(logger *zap.Logger) zapLogger {
	if logger == nil {
		logger = zap.NewNop()
	}
	return zapLogger{logger: logger.Sugar()}
}

func (l zapLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l zapLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l zapLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l zapLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l zapLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}
