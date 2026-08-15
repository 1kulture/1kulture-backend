package logger

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *logrus.Logger

type ContextHook struct{}

func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook ContextHook) Fire(entry *logrus.Entry) error {
	if pc, file, line, ok := runtime.Caller(8); ok {
		funcName := runtime.FuncForPC(pc).Name()
		entry.Data["source"] = fmt.Sprintf("%s:%d:%s", path.Base(file), line, path.Base(funcName))
	}
	return nil
}

func Init(environment string) {
	Log = logrus.New()

	// Set log level based on environment
	if environment == "production" {
		Log.SetLevel(logrus.InfoLevel)
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		})

		// Configure log rotation for production
		Log.SetOutput(&lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    100, // megabytes
			MaxBackups: 30,
			MaxAge:     28, // days
			Compress:   true,
		})
	} else {
		Log.SetLevel(logrus.DebugLevel)
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				return "", fmt.Sprintf("%s:%d", filename, f.Line)
			},
		})
		Log.SetOutput(os.Stdout)
		Log.SetReportCaller(true)
	}

	// Add custom context hook
	Log.AddHook(ContextHook{})
}

// Helper functions for structured logging
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}

func WithError(err error) *logrus.Entry {
	return Log.WithError(err)
}

func WithRequest(c *gin.Context) *logrus.Entry {
	fields := logrus.Fields{
		"request_id": c.GetString("request_id"),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	}

	if userID, exists := c.Get("user_id"); exists {
		fields["user_id"] = userID
	}

	return Log.WithFields(fields)
}

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

func Error(args ...interface{}) {
	Log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

func Debug(args ...interface{}) {
	Log.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

func Warning(args ...interface{}) {
	Log.Warning(args...)
}

func Warningf(format string, args ...interface{}) {
	Log.Warningf(format, args...)
}

func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	Log.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	Log.Panicf(format, args...)
}

// AuditLog creates an audit log entry
func AuditLog(action, resource, resourceID, userID, ip, userAgent string, status string, details map[string]interface{}) {
	fields := logrus.Fields{
		"action":      action,
		"resource":    resource,
		"resource_id": resourceID,
		"user_id":     userID,
		"ip_address":  ip,
		"user_agent":  userAgent,
		"status":      status,
		"timestamp":   time.Now().UTC(),
	}

	for k, v := range details {
		fields[k] = v
	}

	Log.WithFields(fields).Info("AUDIT_LOG")
}
