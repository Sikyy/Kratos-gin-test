package pkg

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Logger struct {
	ZapLogger                 *log.Helper
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	SkipCallerLookup          bool
	IgnoreRecordNotFoundError bool
	Kind                      string
}

func NewZapGormLogV2(zapLogger *log.Helper, kind string) Logger {
	return Logger{
		ZapLogger:                 zapLogger,
		LogLevel:                  gormlogger.Info,
		SlowThreshold:             100 * time.Millisecond,
		SkipCallerLookup:          false,
		IgnoreRecordNotFoundError: false,
		Kind:                      kind,
	}
}

func (l Logger) SetAsDefault() {
	gormlogger.Discard = l
}

func (l Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return Logger{
		ZapLogger:                 l.ZapLogger,
		SlowThreshold:             l.SlowThreshold,
		LogLevel:                  level,
		SkipCallerLookup:          l.SkipCallerLookup,
		IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
	}
}

func (l Logger) Info(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Info {
		return
	}
	l.ZapLogger.WithContext(ctx).Debugf(str, args...)
}

func (l Logger) Warn(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Warn {
		return
	}
	l.ZapLogger.WithContext(ctx).Warnf(str, args...)
}

func (l Logger) Error(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Error {
		return
	}
	l.ZapLogger.WithContext(ctx).Errorf(str, args...)
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= 0 {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		l.ZapLogger.WithContext(ctx).Errorw("kind", l.Kind, "message", err, "elapsed", elapsed, "rows", rows, "sql", sql)
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		l.ZapLogger.WithContext(ctx).Warnw("kind", l.Kind, "message", err, "elapsed", elapsed, "rows", rows, "sql", sql)
	case l.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		l.ZapLogger.WithContext(ctx).Debugw("kind", l.Kind, "message", err, "elapsed", elapsed, "rows", rows, "sql", sql)
	default:
		sql, rows := fc()
		l.ZapLogger.WithContext(ctx).Infow("kind", l.Kind, "message", err, "elapsed", elapsed, "rows", rows, "sql", sql)
	}
}
