package main

import (
	"context"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

func main() {
	config := newConfig()
	alcotest.DoAndExit(config.alcotest)

	ctx := context.Background()

	clients := newClients(ctx, config)
	adapters := newAdapters(config)
	services := newService(config)
	port := newPort(adapters, services)

	go services.server.Start(clients.health.EndCtx(), httputils.Handler(port, clients.health))

	clients.health.WaitForTermination(services.server.Done())

	server.GracefulWait(services.server.Done())
}
