package ui

import (
	"net/http"

	"github.com/slok/go-http-metrics/middleware/std"
)

func (u ui) registerRoutes() {
	u.router.Handle("/static/*", http.StripPrefix(u.servePrefix+"/static", http.FileServer(http.FS(u.staticFS))))

	u.wrapGet("/", u.index())
}

func (u ui) wrapGet(pattern string, h http.HandlerFunc) {
	u.router.With(
		// Add endpoint middlewares.
		std.HandlerProvider(pattern, u.metricsMiddleware),
	).Get(pattern, h)
}
