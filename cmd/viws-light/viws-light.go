package main

import (
	"context"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

func main() {
	config := newConfig()
	alcotest.DoAndExit(config.alcotest)

	ctx := context.Background()

	clients := newClients(ctx, config)
	services := newServices(config)
	port := newPort(clients, services)

	go services.server.Start(clients.health.EndCtx(), port)

	clients.health.WaitForTermination(services.server.Done())
	server.GracefulWait(services.server.Done())
}
