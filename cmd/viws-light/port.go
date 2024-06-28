package main

import (
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

func newPort(clients clients, services services) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /env", model.ChainMiddlewares(services.env.Handler(), services.owasp.Middleware, services.cors.Middleware))
	mux.Handle("GET /", model.ChainMiddlewares(services.viws.Handler(), services.owasp.Middleware, services.cors.Middleware))

	return httputils.Handler(mux, clients.health)
}
