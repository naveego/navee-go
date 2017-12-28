package middleware

import (
	"context"
	"net/http"
)

type Middleware interface {
	WrapHandler(handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error) func(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

type MiddlewareWithVars interface {
	WrapHandler(handler func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error) func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error
}
