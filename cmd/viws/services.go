package main

import (
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/viws/pkg/env"
	"github.com/ViBiOh/viws/pkg/viws"
)

type services struct {
	server *server.Server
	env    env.Service
	viws   viws.App
}

func newService(config configuration) *services {
	return &services{
		env:    env.New(config.env),
		viws:   viws.New(config.viws),
		server: server.New(config.server),
	}
}
