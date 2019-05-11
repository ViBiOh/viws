package main

import (
	"flag"
	"net/http"
	"os"

	httputils "github.com/ViBiOh/httputils/pkg"
	"github.com/ViBiOh/httputils/pkg/alcotest"
	"github.com/ViBiOh/httputils/pkg/cors"
	"github.com/ViBiOh/httputils/pkg/gzip"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/opentracing"
	"github.com/ViBiOh/httputils/pkg/owasp"
	"github.com/ViBiOh/httputils/pkg/prometheus"
	"github.com/ViBiOh/httputils/pkg/server"
	"github.com/ViBiOh/viws/pkg/env"
	"github.com/ViBiOh/viws/pkg/viws"
)

func main() {
	fs := flag.NewFlagSet("viws", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	opentracingConfig := opentracing.Flags(fs, "tracing")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	viwsConfig := viws.Flags(fs, "")
	envConfig := env.Flags(fs, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		logger.Fatal("%+v", err)
	}

	alcotest.DoAndExit(alcotestConfig)

	serverApp, err := httputils.New(serverConfig)
	if err != nil {
		logger.Fatal("%+v", err)
	}

	healthcheckApp := healthcheck.New()
	prometheusApp := prometheus.New(prometheusConfig)
	opentracingApp := opentracing.New(opentracingConfig)
	gzipApp := gzip.New()
	owaspApp := owasp.New(owaspConfig)
	corsApp := cors.New(corsConfig)

	viwsApp, err := viws.New(viwsConfig)
	if err != nil {
		logger.Fatal("%+v", err)
	}
	envApp := env.New(envConfig)

	viwsHandler := server.ChainMiddlewares(viwsApp.Handler(), owaspApp)
	envHandler := server.ChainMiddlewares(envApp.Handler(), owaspApp, corsApp)
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/env" {
			envHandler.ServeHTTP(w, r)
		} else {
			viwsHandler.ServeHTTP(w, r)
		}
	})
	apiHandler := server.ChainMiddlewares(requestHandler, prometheusApp, opentracingApp, gzipApp)

	serverApp.ListenAndServe(apiHandler, nil, healthcheckApp)
}
