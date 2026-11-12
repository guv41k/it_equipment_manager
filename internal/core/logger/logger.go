package core_logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
	file *os.File
}

type loggerContextKey struct{}

func ToContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, log)
}

func FromContext(ctx context.Context) *Logger {
	log, ok := ctx.Value(loggerContextKey{}).(*Logger)
	if !ok {
		panic("no logger in context")
	}
	return log
}

func NewLogger(config Config) (*Logger, error) {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(strings.TrimSpace(config.Level))); err != nil {
		return nil, fmt.Errorf("parse logger level: %w", err)
	}
	if strings.TrimSpace(config.Folder) == "" {
		return nil, fmt.Errorf("logger folder must not be empty")
	}
	if err := os.MkdirAll(config.Folder, 0750); err != nil {
		return nil, fmt.Errorf("create logger folder: %w", err)
	}

	name := time.Now().UTC().Format("2006-01-02T15-04-05.000000000") + ".log"
	file, err := os.OpenFile(filepath.Join(config.Folder, name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open logger file: %w", err)
	}

	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(file), level),
	)

	return &Logger{
		Logger: zap.New(core, zap.AddCaller()),
		file:   file,
	}, nil
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		file:   l.file,
	}
}

func (l *Logger) Close() error {
	return errors.Join(l.file.Sync(), l.file.Close())
}
