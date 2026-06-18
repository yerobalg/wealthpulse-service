package logger

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/yerobalg/wealthpulse-service/helper/appcontext"
	"github.com/yerobalg/wealthpulse-service/helper/authcontext"
)

// Logger wraps a logrus.Logger instance and implements the Interface.
type Logger struct {
	log *logrus.Logger
}

// Interface defines the logging contract for all log levels.
// Each method accepts a context (used to extract request metadata),
// a message string, and optional extra fields to include in the log entry.
type Interface interface {
	Debug(context.Context, string, ...any)
	Info(context.Context, string, ...any)
	Warn(context.Context, string, ...any)
	Error(context.Context, string, ...any)
	Fatal(context.Context, string, ...any)
}

// Init creates and returns a new Logger configured with JSON output and
// automatic caller reporting. The caller location is formatted as the last
// three path segments of the source file and its line number.
func Init() Interface {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fileInfo(8)
		},
	})
	log.SetReportCaller(true)

	return &Logger{
		log: log,
	}
}

// Debug logs a message at the DEBUG level with request metadata from ctx.
func (l *Logger) Debug(ctx context.Context, message string, field ...any) {
	l.log.WithContext(ctx).WithFields(getFields(ctx, field...)).Debug(message)
}

// Info logs a message at the INFO level with request metadata from ctx.
func (l *Logger) Info(ctx context.Context, message string, field ...any) {
	l.log.WithContext(ctx).WithFields(getFields(ctx, field...)).Info(message)
}

// Warn logs a message at the WARN level with request metadata from ctx.
func (l *Logger) Warn(ctx context.Context, message string, field ...any) {
	l.log.WithContext(ctx).WithFields(getFields(ctx, field...)).Warn(message)
}

// Error logs a message at the ERROR level with request metadata from ctx.
func (l *Logger) Error(ctx context.Context, message string, field ...any) {
	l.log.WithContext(ctx).WithFields(getFields(ctx, field...)).Error(message)
}

// Fatal logs a message at the FATAL level with request metadata from ctx,
// then calls os.Exit(1).
func (l *Logger) Fatal(ctx context.Context, message string, field ...any) {
	l.log.WithContext(ctx).WithFields(getFields(ctx, field...)).Fatal(message)
}

func getFields(ctx context.Context, fields ...any) logrus.Fields {
	metadata := appcontext.GetMetadata(ctx)
	logFields := logrus.Fields{
		"request_id":      metadata.RequestID,
		"service_version": metadata.ServiceVersion,
		"user_agent":      metadata.UserAgent,
		"device_type":     metadata.DeviceType,
		"source_ip":       metadata.SourceIP,
		"user_id":         authcontext.GetUserID(ctx),
	}

	if len(fields) > 0 {
		logFields["data"] = fields[0]
	} else {
		logFields["data"] = nil
	}

	return logFields
}

func fileInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "<???>"
	} else {
		location := strings.Split(file, "/")
		file = strings.Join(location[len(location)-3:], "/")
	}
	return fmt.Sprintf("%s:%d", file, line)
}
