package main

import (
	"context"
	"flag"
	"net/http"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/pprof"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/viws/pkg/env"
	"github.com/ViBiOh/viws/pkg/viws"
	"github.com/klauspost/compress/gzhttp"
)

func main() {
	fs := flag.NewFlagSet("viws", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	appServerConfig := server.Flags(fs, "")
	healthConfig := health.Flags(fs, "")

	alcotestConfig := alcotest.Flags(fs, "")
	loggerConfig := logger.Flags(fs, "logger")
	telemetryConfig := telemetry.Flags(fs, "telemetry")
	pprofConfig := pprof.Flags(fs, "pprof")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	gzip := flags.New("Gzip", "Enable gzip compression").DocPrefix("gzip").Bool(fs, true, nil)

	viwsConfig := viws.Flags(fs, "")
	envConfig := env.Flags(fs, "")

	_ = fs.Parse(os.Args[1:])

	alcotest.DoAndExit(alcotestConfig)

	logger.Init(loggerConfig)

	ctx := context.Background()

	healthService := health.New(ctx, healthConfig)

	telemetryApp, err := telemetry.New(ctx, telemetryConfig)
	logger.FatalfOnErr(ctx, err, "create telemetry")

	defer telemetryApp.Close(ctx)

	logger.AddOpenTelemetryToDefaultLogger(telemetryApp)

	service, version, envName := telemetryApp.GetServiceVersionAndEnv()
	pprofApp := pprof.New(pprofConfig, service, version, envName)

	go pprofApp.Start(healthService.DoneCtx())

	appServer := server.New(appServerConfig)

	owaspApp := owasp.New(owaspConfig)
	corsApp := cors.New(corsConfig)

	envApp := env.New(envConfig)
	viwsApp := viws.New(viwsConfig)

	viwsHandler := model.ChainMiddlewares(viwsApp.Handler(), owaspApp.Middleware)
	envHandler := model.ChainMiddlewares(envApp.Handler(), owaspApp.Middleware, corsApp.Middleware)
	appHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/env" {
			telemetry.SetRouteTag(r.Context(), "/env")
			envHandler.ServeHTTP(w, r)
		} else {
			viwsHandler.ServeHTTP(w, r)
		}
	})

	middlewares := []model.Middleware{recoverer.Middleware, telemetryApp.Middleware("http")}
	if *gzip {
		middlewares = append(middlewares, func(next http.Handler) http.Handler {
			return gzhttp.GzipHandler(next)
		})
	}

	go appServer.Start(healthService.EndCtx(), httputils.Handler(appHandler, healthService, middlewares...))

	healthService.WaitForTermination(appServer.Done())

	server.GracefulWait(appServer.Done())
}
