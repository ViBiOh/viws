package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/httputils/v3/pkg/alcotest"
	"github.com/ViBiOh/httputils/v3/pkg/cors"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/owasp"
	"github.com/ViBiOh/httputils/v3/pkg/prometheus"
	"github.com/ViBiOh/viws/pkg/env"
	"github.com/ViBiOh/viws/pkg/viws"
)

func main() {
	fs := flag.NewFlagSet("viws", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	viwsConfig := viws.Flags(fs, "")
	envConfig := env.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)

	owaspApp := owasp.New(owaspConfig)
	corsApp := cors.New(corsConfig)

	viwsApp, err := viws.New(viwsConfig)
	logger.Fatal(err)
	envApp := env.New(envConfig)

	viwsHandler := httputils.ChainMiddlewares(viwsApp.Handler(), owaspApp.Middleware)
	envHandler := httputils.ChainMiddlewares(envApp.Handler(), owaspApp.Middleware, corsApp.Middleware)
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/env" {
			envHandler.ServeHTTP(w, r)
		} else {
			viwsHandler.ServeHTTP(w, r)
		}
	})

	server := httputils.New(serverConfig)
	server.Middleware(prometheus.New(prometheusConfig).Middleware)
	server.Middleware(gziphandler.GzipHandler)
	server.ListenServeWait(requestHandler)
}
