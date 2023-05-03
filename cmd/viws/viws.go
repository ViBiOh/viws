package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/prometheus"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/ViBiOh/viws/pkg/env"
	"github.com/ViBiOh/viws/pkg/viws"
	"github.com/klauspost/compress/gzhttp"
)

func main() {
	fs := flag.NewFlagSet("viws", flag.ExitOnError)

	appServerConfig := server.Flags(fs, "")
	promServerConfig := server.Flags(fs, "prometheus", flags.NewOverride("Port", uint(9090)), flags.NewOverride("IdleTimeout", 10*time.Second), flags.NewOverride("ShutdownTimeout", 5*time.Second))
	healthConfig := health.Flags(fs, "")

	alcotestConfig := alcotest.Flags(fs, "")
	loggerConfig := logger.Flags(fs, "logger")
	tracerConfig := tracer.Flags(fs, "tracer")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	gzip := flags.New("Gzip", "Enable gzip compression").DocPrefix("gzip").Bool(fs, true, nil)

	viwsConfig := viws.Flags(fs, "")
	envConfig := env.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)
	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	ctx := context.Background()

	tracerApp, err := tracer.New(ctx, tracerConfig)
	logger.Fatal(err)
	defer tracerApp.Close(ctx)
	request.AddTracerToDefaultClient(tracerApp.GetProvider())

	go func() {
		fmt.Println(http.ListenAndServe("localhost:9999", http.DefaultServeMux))
	}()

	appServer := server.New(appServerConfig)
	promServer := server.New(promServerConfig)
	prometheusApp := prometheus.New(prometheusConfig)
	healthApp := health.New(healthConfig)

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

	middlewares := []model.Middleware{recoverer.Middleware, prometheusApp.Middleware, tracerApp.Middleware}
	if *gzip {
		middlewares = append(middlewares, func(next http.Handler) http.Handler {
			return gzhttp.GzipHandler(next)
		})
	}

	endCtx := healthApp.End(ctx)

	go promServer.Start(endCtx, "prometheus", prometheusApp.Handler())
	go appServer.Start(endCtx, "http", httputils.Handler(appHandler, healthApp, middlewares...))

	healthApp.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done(), promServer.Done())
}
