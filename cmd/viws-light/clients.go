package main

import (
	"context"

	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

type clients struct {
	health *health.Service
}

func newClients(ctx context.Context, config configuration) clients {
	var output clients

	logger.Init(ctx, config.logger)

	output.health = health.New(ctx, config.health)

	return output
}
