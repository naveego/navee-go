package logging

import (
	"context"

	"github.com/sirupsen/logrus"
)

type contextKey string

var (
	loggerKey = contextKey("logger")
)

func (c contextKey) String() string {
	return "logging:" + string(c)
}

// SetLogger set up the log entry on the given context
func SetLogger(ctx context.Context, log *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

// GetLogger gets the log entry from the given context.
// If no logger is set, it will return a default logger.
func GetLogger(ctx context.Context) *logrus.Entry {
	e := ctx.Value(loggerKey)
	if e == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	l, ok := e.(*logrus.Entry)
	if !ok {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return l
}
