package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

// Chain composes middlewares around a final handler.
func Chain(
	h http.Handler,
	middlewares ...Middleware,
) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}
