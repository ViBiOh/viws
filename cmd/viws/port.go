package main

import (
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/model"
)

func newPort(adapters adapters, services *services) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /env", model.ChainMiddlewares(services.env.Handler(), adapters.owasp.Middleware, adapters.cors.Middleware))
	mux.Handle("GET /", model.ChainMiddlewares(services.viws.Handler(), adapters.owasp.Middleware, adapters.cors.Middleware))

	return mux
}
