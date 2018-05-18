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

func main() {
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)
	viwsConfig := viws.Flags(``)
	envConfig := env.Flags(``)
	opentracingConfig := opentracing.Flags(`tracing`)

	healthcheckApp := healthcheck.NewApp()

	httputils.NewApp(httputils.Flags(``), func() http.Handler {
		envApp := env.NewApp(envConfig)
		viwsApp, err := viws.NewApp(viwsConfig)
		if err != nil {
			log.Fatalf(`Error while instanciating viws: %v`, err)
		}

		viwsHandler := owasp.Handler(owaspConfig, viwsApp.Handler())
		envHandler := owasp.Handler(owaspConfig, cors.Handler(corsConfig, envApp.Handler()))
		requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == `/env` {
				envHandler.ServeHTTP(w, r)
			} else {
				viwsHandler.ServeHTTP(w, r)
			}
		})

		apiHandler := opentracing.NewApp(opentracingConfig).Handler(gziphandler.GzipHandler(requestHandler))
		healthcheckHandler := healthcheckApp.Handler(nil)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == `/health` {
				healthcheckHandler.ServeHTTP(w, r)
			} else {
				apiHandler.ServeHTTP(w, r)
			}
		})
	}, nil, healthcheckApp).ListenAndServe()
}
