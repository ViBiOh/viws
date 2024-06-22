package main

import (
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
)

type adapters struct {
	cors  cors.Service
	owasp owasp.Service
}

func newAdapters(config configuration) adapters {
	return adapters{
		owasp: owasp.New(config.owasp),
		cors:  cors.New(config.cors),
	}
}
