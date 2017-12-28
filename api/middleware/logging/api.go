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
	getRouteFunc  func(ctx context.Context) string
	getTenantFunc func(ctx context.Context) string
}

// NewAPILoggingMiddleware returns a new instance of the API Logging Middleware.
func NewAPILoggingMiddleware(getRouteFunc, getTenantFunc func(ctx context.Context) string) middleware.Middleware {
	return &apiLoggingMiddleware{getRouteFunc, getTenantFunc}
}

func (l *apiLoggingMiddleware) WrapHandler(handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error) func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return logAndExecute(ctx, w, r, l.getRouteFunc, l.getTenantFunc, handler)
	}
}

type apiLoggingMiddlewareWithVars struct {
	*apiLoggingMiddleware
}

func NewAPILoggingMiddlewareWithVars(getRouteFunc, getTenantFunc func(ctx context.Context) string) middleware.MiddlewareWithVars {
	return &apiLoggingMiddlewareWithVars{
		&apiLoggingMiddleware{
			getRouteFunc,
			getTenantFunc,
		},
	}
}

func (l *apiLoggingMiddlewareWithVars) WrapHandler(handler func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error) func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
		return logAndExecute(ctx, w, r, l.getRouteFunc, l.getTenantFunc, func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return handler(ctx, w, r, map[string]string{})
		})
	}
}

func logAndExecute(ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	routeFunc func(ctx context.Context) string,
	tenantFunc func(ctx context.Context) string,
	handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error) error {

	// Create the logger entry with correlation id
	logger := logrus.WithField("correlation_id", strings.ToLower(uuid.NewV4().String()))

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

	tenant := tenantFunc(ctx)
	routePath := routeFunc(ctx)

	if routePath != "" {
		requestDurationsHistogram.WithLabelValues(r.Method, routePath).Observe(elapsed.Seconds())
		requestBytesHistogram.WithLabelValues(r.Method, routePath).Observe(float64(bytesIn))
		responseBytesHistogram.WithLabelValues(r.Method, routePath).Observe(float64(bytesOut))
	}

	// create the log entry
	logEntry := logger.WithFields(logrus.Fields{
		"tenant":          tenant,
		"api_route":       routePath,
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
