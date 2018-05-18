package main

import (
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/httputils/pkg"
	"github.com/ViBiOh/httputils/pkg/cors"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/opentracing"
	"github.com/ViBiOh/httputils/pkg/owasp"
	"github.com/ViBiOh/viws/pkg/env"
	"github.com/ViBiOh/viws/pkg/viws"
)

const (
	healthPrefix = `/health`
	envPrefix    = `/env`
)

func main() {
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)
	viwsConfig := viws.Flags(``)
	envConfig := env.Flags(``)
	opentracingConfig := opentracing.Flags(`tracing`)

	healthcheckApp := healthcheck.NewApp()

	httputils.NewApp(httputils.Flags(``), func() http.Handler {
		viwsApp, err := viws.NewApp(viwsConfig)
		if err != nil {
			log.Fatalf(`Error while instanciating viws: %v`, err)
		}

		envApp := env.NewApp(envConfig)

		requestsHandler := owasp.Handler(owaspConfig, viwsApp.Handler())
		envHandler := owasp.Handler(owaspConfig, cors.Handler(corsConfig, envApp.Handler()))
		healthcheckHandler := healthcheckApp.Handler(nil)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == `/health` {
				healthcheckHandler.ServeHTTP(w, r)
			} else if r.URL.Path == envPrefix {
				envHandler.ServeHTTP(w, r)
			} else {
				requestsHandler.ServeHTTP(w, r)
			}
		})

		return opentracing.NewApp(opentracingConfig).Handler(gziphandler.GzipHandler(handler))
	}, nil, healthcheckApp).ListenAndServe()
}
