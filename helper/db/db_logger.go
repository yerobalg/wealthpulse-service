package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm/logger"

	appLogger "github.com/yerobalg/wealthpulse-service/helper/logger"
)

// ErrRecordNotFound record not found error
var ErrRecordNotFound = errors.New("record not found")

// LogLevel log level
type LogLevel int

const (
	// Silent silent log level
	Silent LogLevel = iota + 1
	// Error error log level
	Error
	// Warn warn log level
	Warn
	// Info info log level
	Info
)

// Config logger config
type Config struct {
	SlowThreshold             time.Duration
	Colorful                  bool
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
	LogLevel                  LogLevel
}

// Interface logger interface
type GormLoggerInterface interface {
	LogMode(logger.LogLevel) logger.Interface
	Info(context.Context, string, ...any)
	Warn(context.Context, string, ...any)
	Error(context.Context, string, ...any)
	Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error)
}

// New initialize logger
func InitGormLogger(log appLogger.Interface) GormLoggerInterface {
	return &GormLogger{
		Config: Config{
			LogLevel:             Info,
			ParameterizedQueries: true,
		},
		AppLogger: log,
	}
}

type GormLogger struct {
	Config
	AppLogger appLogger.Interface
}

type SQLLogger struct {
	TimeElapsed  string `json:"time_elapsed"`
	RowsAffected int64  `json:"rows_affected"`
	Query        string `json:"query"`
	IsError      bool   `json:"is_error"`
}

// LogMode log mode
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info print info
func (l GormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= Info {
		l.AppLogger.Info(ctx, msg, data...)
	}
}

// Warn print warn messages
func (l GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= Warn {
		l.AppLogger.Warn(ctx, msg, data...)
	}
}

// Error print error messages
func (l GormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= Error {
		l.AppLogger.Error(ctx, msg, data...)
	}
}

// Trace print sql message
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= Silent {
		return
	}

	sql, rows := fc()
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= Error:
		sqlLogger := SQLLogger{
			TimeElapsed:  fmt.Sprintf("%.2fms", float64(elapsed.Nanoseconds())/1e6),
			RowsAffected: rows,
			Query:        sql,
		}
		if rows == -1 {
			sqlLogger.RowsAffected = 0
		}

		l.Error(ctx, err.Error(), sqlLogger)
	case l.LogLevel == Info:
		sqlLogger := SQLLogger{
			TimeElapsed:  fmt.Sprintf("%.2fms", float64(elapsed.Nanoseconds())/1e6),
			RowsAffected: rows,
			Query:        sql,
		}
		if rows == -1 {
			sqlLogger.RowsAffected = 0
		}

		l.Info(ctx, "sql", sqlLogger)
	}
}

func (l GormLogger) ParamsFilter(ctx context.Context, sql string, params ...any) (string, []any) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
