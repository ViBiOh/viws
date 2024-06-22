package main

import (
	"context"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/klauspost/compress/gzhttp"
)

func main() {
	config := newConfig()
	alcotest.DoAndExit(config.alcotest)

	ctx := context.Background()

	clients, err := newClients(ctx, config)
	logger.FatalfOnErr(ctx, err, "clients")

	defer clients.Close(ctx)
	go clients.Start()

	adapters := newAdapters(config)
	services := newService(config)
	port := newPort(adapters, services)

	middlewares := []model.Middleware{clients.telemetry.Middleware("http")}
	if *config.gzip {
		middlewares = append(middlewares, func(next http.Handler) http.Handler {
			return gzhttp.GzipHandler(next)
		})
	}

	go services.server.Start(clients.health.EndCtx(), httputils.Handler(port, clients.health, middlewares...))

	clients.health.WaitForTermination(services.server.Done())

	server.GracefulWait(services.server.Done())
}
