package main

import (
	"flag"
	"net/http"

	httputils "github.com/ViBiOh/httputils/pkg"
	"github.com/ViBiOh/httputils/pkg/alcotest"
	"github.com/ViBiOh/httputils/pkg/cors"
	"github.com/ViBiOh/httputils/pkg/gzip"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/owasp"
	"github.com/ViBiOh/httputils/pkg/server"
	"github.com/ViBiOh/viws/pkg/env"
	"github.com/ViBiOh/viws/pkg/viws"
)

func main() {
	serverConfig := httputils.Flags(``)
	alcotestConfig := alcotest.Flags(``)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)

	viwsConfig := viws.Flags(``)
	envConfig := env.Flags(``)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	serverApp := httputils.NewApp(serverConfig)
	healthcheckApp := healthcheck.NewApp()
	gzipApp := gzip.NewApp()
	owaspApp := owasp.NewApp(owaspConfig)
	corsApp := cors.NewApp(corsConfig)

	viwsApp, err := viws.NewApp(viwsConfig)
	if err != nil {
		logger.Error(`%+v`, err)
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
	apiHandler := server.ChainMiddlewares(requestHandler, gzipApp)

	serverApp.ListenAndServe(apiHandler, nil, healthcheckApp)
}
