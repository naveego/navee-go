package logging

import (
	"context"

	"github.com/Sirupsen/logrus"
)

// Options defines the logging middleware options.
type Options struct {
	loggerFactory func(ctx context.Context) *logrus.Entry
}

// Option configures the logging middleware.
type Option func(*Options)

// WithLoggerFactory configures the logging middleware with a method of creating
// the base logger entry.  This can be useful when the system using the middleware
// needs to inject fields in addition to the api fields already logged.
func WithLoggerFactory(factory func(ctx context.Context) *logrus.Entry) Option {
	return func(o *Options) {
		o.loggerFactory = factory
	}
}
