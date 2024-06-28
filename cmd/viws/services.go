package main

import (
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/viws/pkg/env"
	"github.com/ViBiOh/viws/pkg/viws"
)

type services struct {
	server *server.Server
	cors   cors.Service
	owasp  owasp.Service

	env  env.Service
	viws viws.App
}

func newServices(config configuration) services {
	var output services

	output.server = server.New(config.server)
	output.owasp = owasp.New(config.owasp)
	output.cors = cors.New(config.cors)

	output.env = env.New(config.env)
	output.viws = viws.New(config.viws)

	return output
}
