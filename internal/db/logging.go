package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	slogger "github.com/act3-ai/go-common/pkg/logger"
)

// NewGormLogger creates a GORM compatible logger that actually logs to the given logr.Logger

// Adapted from https://github.com/vchitai/logrgorm2/blob/main/logrgorm.go

// GormLogger is an interface that also implements the gorm logger interface (with a few more functions).
type GormLogger interface {
	logger.Interface
	IgnoreRecordNotFoundError(bool) GormLogger
	SlowThreshold(time.Duration) GormLogger
}

// gormLoggerImpl implements the GormLogger interface.
type gormLoggerImpl struct {
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

// NewGormLogger creates a GORM compatible logger that actually logs to the given logr.Logger.
func NewGormLogger() GormLogger {
	return &gormLoggerImpl{
		slowThreshold:             100 * time.Millisecond,
		ignoreRecordNotFoundError: true,
	}
}

func (l *gormLoggerImpl) IgnoreRecordNotFoundError(ignore bool) GormLogger {
	return &gormLoggerImpl{
		slowThreshold:             l.slowThreshold,
		ignoreRecordNotFoundError: ignore,
	}
}

func (l *gormLoggerImpl) SlowThreshold(threshold time.Duration) GormLogger {
	return &gormLoggerImpl{
		slowThreshold:             threshold,
		ignoreRecordNotFoundError: l.ignoreRecordNotFoundError,
	}
}

func (l *gormLoggerImpl) LogMode(level logger.LogLevel) logger.Interface {
	// This is only called by con.Debug() by gorm.
	// logr.Logger.V() does not let you lower the verbosity level, you can only increase it.
	panic("Calling LogMode() has no effect.  Did you really mean to call it? (probably via con.Debug())")
	// NOOP
	// return l
	/*
		return &gormLoggerImpl{
			ll:                        l.ll.V(int(logger.Warn - level)),
			slowThreshold:             l.slowThreshold,
			ignoreRecordNotFoundError: l.ignoreRecordNotFoundError,
		}
	*/

	// To implement this we would need to have a logger with the level set to one lower.  Then we can increase all of them by an adjustment (level passed in).
}

func (l *gormLoggerImpl) Info(ctx context.Context, s string, i ...any) {
	log := slogger.FromContext(ctx)
	if !log.Enabled(ctx, slog.LevelInfo) {
		return
	}
	log.InfoContext(ctx, "Database info", "wrapped", fmt.Sprintf(s, i...))
}

func (l *gormLoggerImpl) Warn(ctx context.Context, s string, i ...any) {
	log := slogger.FromContext(ctx)
	if !log.Enabled(ctx, slog.LevelInfo) {
		return
	}
	log.InfoContext(ctx, "Database warning", "wrapped", fmt.Sprintf(s, i...))
}

func (l *gormLoggerImpl) Error(ctx context.Context, s string, i ...any) {
	log := slogger.FromContext(ctx)
	log.ErrorContext(ctx, "Database error", "wrapped", fmt.Sprintf(s, i...))
}

func (l *gormLoggerImpl) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	log := slogger.FromContext(ctx)
	elapsed := time.Since(begin)
	switch {
	case err != nil && (!l.ignoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		log.ErrorContext(ctx, "SQL Error", "elapsed", elapsed, "rows", rows, "sql", sql, "error", err)
	case l.slowThreshold != 0 && elapsed > l.slowThreshold && log.Enabled(ctx, slog.LevelInfo):
		sql, rows := fc()
		log.InfoContext(ctx, "Slow SQL", "elapsed", elapsed, "rows", rows, "sql", sql)
	case log.Enabled(ctx, slog.LevelDebug):
		sql, rows := fc()
		log.DebugContext(ctx, "SQL Executed", "elapsed", elapsed, "rows", rows, "sql", sql)
	}
}
