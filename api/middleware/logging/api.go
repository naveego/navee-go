package logging

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/naveego/navee-go/api/middleware"
	"github.com/naveego/navee-go/logging"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type apiLoggingMiddleware struct {
	options *Options
}

// NewAPILoggingMiddleware returns a new instance of the API Logging Middleware.
func NewAPILoggingMiddleware(opt ...Option) middleware.Middleware {
	opts := new(Options)
	for _, f := range opt {
		f(opts)
	}

	if opts.loggerFactory == nil {
		opts.loggerFactory = func(ctx context.Context) *logrus.Entry {
			return logrus.NewEntry(logrus.StandardLogger())
		}
	}

	return &apiLoggingMiddleware{opts}
}

func (l *apiLoggingMiddleware) WrapHandler(handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error) func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return logAndExecute(ctx, w, r, l.options.loggerFactory, handler)
	}
}

type apiLoggingMiddlewareWithVars struct {
	*apiLoggingMiddleware
}

// NewAPILoggingMiddlewareWithVars returns a new instance of the API Logging Middleware
// but includes the vars map[string]string parameter to support some of our RESTful
// server code.
func NewAPILoggingMiddlewareWithVars(opt ...Option) middleware.MiddlewareWithVars {
	return &apiLoggingMiddlewareWithVars{NewAPILoggingMiddleware(opt)}
}

func (l *apiLoggingMiddlewareWithVars) WrapHandler(handler func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error) func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
		return logAndExecute(ctx, w, r, l.getLogger, func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return handler(ctx, w, r, map[string]string{})
		})
	}
}

func logAndExecute(ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	loggerFactory func(ctx context.Context) *logrus.Entry,
	handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error) error {

	// Create the logger entry with correlation id
	logger := loggerFactory(ctx).WithField("correlation_id", strings.ToLower(uuid.NewV4().String()))

	start := time.Now()

	loggingWriter := newLoggingResponseWriter(w)
	handlerCtx := logging.SetLogger(ctx, logger)
	err := handler(handlerCtx, loggingWriter, r)

	// determine the elapsed time for the call
	elapsed := time.Since(start)
	elapsedMs := int(math.Floor(float64(elapsed / time.Millisecond)))

	bytesIn, _ := strconv.Atoi(r.Header.Get("Content-Length"))
	bytesOut := loggingWriter.ContentLength
	httpStatus := loggingWriter.StatusCode

	if httpStatus == 0 {
		httpStatus = 200
	}

	routePath, ok := logger.Data["api_route"].(string)
	if ok {
		requestDurationsHistogram.WithLabelValues(r.Method, routePath).Observe(elapsed.Seconds())
		requestBytesHistogram.WithLabelValues(r.Method, routePath).Observe(float64(bytesIn))
		responseBytesHistogram.WithLabelValues(r.Method, routePath).Observe(float64(bytesOut))
	}

	// create the log entry
	logEntry := logger.WithFields(logrus.Fields{
		"api_http_status": httpStatus,
		"api_http_method": r.Method,
		"api_http_path":   r.URL.Path,
		"execution_ms":    elapsedMs,
		"net_bytes_in":    bytesIn,
		"net_bytes_out":   bytesOut,
		"net_proto":       "http",
	})

	if err == nil {
		logEntry.Info("Processed API request")
	}

	return err
}
