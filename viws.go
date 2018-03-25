package main

import (
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/healthcheck"
	"github.com/ViBiOh/httputils/owasp"
	"github.com/ViBiOh/viws/env"
	"github.com/ViBiOh/viws/viws"
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

	httputils.NewApp(httputils.Flags(``), func() http.Handler {
		viwsApp, err := viws.NewApp(viwsConfig)
		if err != nil {
			log.Fatalf(`Error while instanciating viws: %v`, err)
		}

		envApp := env.NewApp(envConfig)

		healthcheckHandler := healthcheck.Handler()
		requestsHandler := owasp.Handler(owaspConfig, viwsApp.Handler())
		envHandler := owasp.Handler(owaspConfig, cors.Handler(corsConfig, envApp.Handler()))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == healthPrefix {
				healthcheckHandler.ServeHTTP(w, r)
			} else if r.URL.Path == envPrefix {
				envHandler.ServeHTTP(w, r)
			} else {
				requestsHandler.ServeHTTP(w, r)
			}
		})

		return gziphandler.GzipHandler(handler)
	}, nil).ListenAndServe()
}
