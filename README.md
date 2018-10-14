# viws

A superlight HTTP fileserver with customizable behavior.

[![Build Status](https://travis-ci.org/ViBiOh/viws.svg?branch=master)](https://travis-ci.org/ViBiOh/viws)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/viws)](https://goreportcard.com/report/github.com/ViBiOh/viws)

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FViBiOh%2Fviws.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FViBiOh%2Fviws?ref=badge_large)

## Installation

```bash
go get github.com/ViBiOh/viws/cmd/viws
```

## Features

* Full TLS support
* Opentracing with Jaeger
* Prometheus monitoring
* Rollbar error reporting
* Read-only container
* Serve static content, with Single Page App handling
* Custom 404 page

## Usage

By default, server is listening on the `1080` port and serve content for GET requests from the `/www/` directory, which have to contains an `index.html`. It assumes that HTTPS is done, somewhere between browser and server (e.g. CloudFlare, ReverseProxy, Traefik, self-signed, ...) so it sets HSTS flag by default, security matters.

```bash
Usage of viws:
  -corsCredentials
      [cors] Access-Control-Allow-Credentials
  -corsExpose string
      [cors] Access-Control-Expose-Headers
  -corsHeaders string
      [cors] Access-Control-Allow-Headers (default "Content-Type")
  -corsMethods string
      [cors] Access-Control-Allow-Methods (default "GET")
  -corsOrigin string
      [cors] Access-Control-Allow-Origin (default "*")
  -csp string
      [owasp] Content-Security-Policy (default "default-src 'self'; base-uri 'self'")
  -directory string
      [viws] Directory to serve (default "/www/")
  -env string
      [env] Environments key variables to expose, comma separated
  -frameOptions string
      [owasp] X-Frame-Options (default "deny")
  -headers string
      [viws] Custom headers, tilde separated (e.g. content-language:fr~X-UA-Compatible:test)
  -hsts
      [owasp] Indicate Strict Transport Security (default true)
  -notFound
      [viws] Graceful 404 page at /404.html
  -port int
      Listen port (default 1080)
  -push string
      [viws] Paths for HTTP/2 Server Push on index, comma separated
  -rollbarEnv string
      [rollbar] Environment (default "prod")
  -rollbarServerRoot string
      [rollbar] Server Root
  -rollbarToken string
      [rollbar] Token
  -spa
      [viws] Indicate Single Page Application mode
  -tls
      Serve TLS content (default true)
  -tlsCert string
      [tls] PEM Certificate file
  -tlsHosts string
      [tls] Self-signed certificate hosts, comma separated (default "localhost")
  -tlsKey string
      [tls] PEM Key file
  -tlsOrganization string
      [tls] Self-signed certificate organization (default "ViBiOh")
  -tracingAgent string
      [opentracing] Jaeger Agent (e.g. host:port) (default "jaeger:6831")
  -tracingName string
      [opentracing] Service name
  -url string
      [health] URL to check
  -userAgent string
      [health] User-Agent used (default "Golang alcotest")
```

## Single Page Application

This mode is useful when you have a router in your javascript framework (e.g. Angular/React/Vue). When a request target a not found file, it returns the index instead of 404. This option also deactivates cache for the index in order to make work the cache-buster for javascript/style files.

e.g.
```bash
curl myWebsite.com/users/vibioh/
=> /index.html
```

Be careful, `-notFound` and `-spa` are incompatible flags. If you set both, you'll get an error.


## Docker

`docker run -d -p 1080:1080 -v /var/www/html/:/www/ vibioh/viws`

We recommend using a Dockerfile to ship your files inside it.

e.g.
```
FROM vibioh/viws

COPY dist/ /www/
```

## Compilation

You need Go 1.9 in order to compile the project.

```bash
make go
```
