package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "net/http/pprof"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
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
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	gzip := flags.New("Gzip", "Enable gzip compression").DocPrefix("gzip").Bool(fs, true, nil)

	viwsConfig := viws.Flags(fs, "")
	envConfig := env.Flags(fs, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	alcotest.DoAndExit(alcotestConfig)

	logger.Init(loggerConfig)

	ctx := context.Background()

	telemetryApp, err := telemetry.New(ctx, telemetryConfig)
	if err != nil {
		slog.Error("create telemetry", "err", err)
		os.Exit(1)
	}

	defer telemetryApp.Close(ctx)

	logger.AddOpenTelemetryToDefaultLogger()

	go func() {
		fmt.Println(http.ListenAndServe("localhost:9999", http.DefaultServeMux))
	}()

	appServer := server.New(appServerConfig)
	healthApp := health.New(ctx, healthConfig)

	owaspApp := owasp.New(owaspConfig)
	corsApp := cors.New(corsConfig)

	envApp := env.New(envConfig)
	viwsApp := viws.New(viwsConfig)

	viwsHandler := model.ChainMiddlewares(viwsApp.Handler(), owaspApp.Middleware)
	envHandler := model.ChainMiddlewares(envApp.Handler(), owaspApp.Middleware, corsApp.Middleware)
	appHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/env" {
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

	go appServer.Start(healthApp.EndCtx(), httputils.Handler(appHandler, healthApp, middlewares...))

	healthApp.WaitForTermination(appServer.Done())

	server.GracefulWait(appServer.Done())
}
