package main

import (
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/klauspost/compress/gzhttp"
)

func newPort(config configuration, clients clients, services services) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /env", model.ChainMiddlewares(services.env.Handler(), services.owasp.Middleware, services.cors.Middleware))
	mux.Handle("GET /", model.ChainMiddlewares(services.viws.Handler(), services.owasp.Middleware, services.cors.Middleware))

	middlewares := []model.Middleware{clients.telemetry.Middleware("http")}
	if *config.gzip {
		middlewares = append(middlewares, func(next http.Handler) http.Handler {
			return gzhttp.GzipHandler(next)
		})
	}

	return httputils.Handler(mux, clients.health, middlewares...)
}
