package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/pprof"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/viws/pkg/env"
	"github.com/ViBiOh/viws/pkg/viws"
)

type configuration struct {
	alcotest  *alcotest.Config
	health    *health.Config
	logger    *logger.Config
	telemetry *telemetry.Config
	pprof     *pprof.Config
	server    *server.Config
	owasp     *owasp.Config
	cors      *cors.Config

	viws *viws.Config
	env  *env.Config
	gzip *bool
}

func newConfig() configuration {
	fs := flag.NewFlagSet("viws", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	config := configuration{
		health:    health.Flags(fs, ""),
		alcotest:  alcotest.Flags(fs, ""),
		logger:    logger.Flags(fs, "logger"),
		telemetry: telemetry.Flags(fs, "telemetry"),
		pprof:     pprof.Flags(fs, "pprof"),
		server:    server.Flags(fs, ""),
		owasp:     owasp.Flags(fs, ""),
		cors:      cors.Flags(fs, "cors"),

		viws: viws.Flags(fs, ""),
		env:  env.Flags(fs, ""),
		gzip: flags.New("Gzip", "Enable gzip compression").DocPrefix("gzip").Bool(fs, true, nil),
	}

	_ = fs.Parse(os.Args[1:])

	return config
}
