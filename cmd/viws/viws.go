package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ViBiOh/httputils/pkg"
	"github.com/ViBiOh/httputils/pkg/alcotest"
	"github.com/ViBiOh/httputils/pkg/cors"
	"github.com/ViBiOh/httputils/pkg/gzip"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/opentracing"
	"github.com/ViBiOh/httputils/pkg/owasp"
	"github.com/ViBiOh/httputils/pkg/prometheus"
	"github.com/ViBiOh/httputils/pkg/rollbar"
	"github.com/ViBiOh/httputils/pkg/server"
	"github.com/ViBiOh/viws/pkg/env"
	"github.com/ViBiOh/viws/pkg/viws"
)

func main() {
	serverConfig := httputils.Flags(``)
	alcotestConfig := alcotest.Flags(``)
	opentracingConfig := opentracing.Flags(`tracing`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)
	prometheusConfig := prometheus.Flags(`prometheus`)
	rollbarConfig := rollbar.Flags(`rollbar`)

	viwsConfig := viws.Flags(``)
	envConfig := env.Flags(``)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	serverApp := httputils.NewApp(serverConfig)
	healthcheckApp := healthcheck.NewApp()
	opentracingApp := opentracing.NewApp(opentracingConfig)
	owaspApp := owasp.NewApp(owaspConfig)
	corsApp := cors.NewApp(corsConfig)
	prometheusApp := prometheus.NewApp(prometheusConfig)
	rollbarApp := rollbar.NewApp(rollbarConfig)
	gzipApp := gzip.NewApp()

	viwsApp, err := viws.NewApp(viwsConfig)
	if err != nil {
		log.Fatalf(`Error while instanciating viws: %v`, err)
	}
	envApp := env.NewApp(envConfig)

	viwsHandler := server.ChainMiddlewares(viwsApp.Handler(), owaspApp)
	envHandler := server.ChainMiddlewares(envApp.Handler(), owaspApp, corsApp)
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == `/env` {
			envHandler.ServeHTTP(w, r)
		} else {
			viwsHandler.ServeHTTP(w, r)
		}
	})
	apiHandler := server.ChainMiddlewares(requestHandler, prometheusApp, opentracingApp, rollbarApp, gzipApp)

	serverApp.ListenAndServe(apiHandler, nil, healthcheckApp, rollbarApp)
}
