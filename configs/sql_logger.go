package configs

import (
	"context"
	"gorm.io/gorm/logger"
	"log/slog"
	"time"
)

// SqlLogger 使用已配置的全局 slog logger
type SqlLogger struct {
	logLevel logger.LogLevel
}

// LogMode 设置日志级别
func (l *SqlLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info 实现 logger.Interface 的 Info 方法
func (l *SqlLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		slog.InfoContext(ctx, msg, "data", data)
	}
}

// Warn 实现 logger.Interface 的 Warn 方法
func (l *SqlLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		slog.WarnContext(ctx, msg, "data", data)
	}
}

// Error 实现 logger.Interface 的 Error 方法
func (l *SqlLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		slog.ErrorContext(ctx, msg, "data", data)
	}
}

// Trace 实现 logger.Interface 的 Trace 方法
func (l *SqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		slog.ErrorContext(ctx, "SQL execution failed",
			"err", err, "sql", sql, "rows", rows, "elapsed", elapsed)
	} else {
		slog.DebugContext(ctx, "SQL executed",
			"sql", sql, "rows", rows, "elapsed", elapsed)
	}
}
