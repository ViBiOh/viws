# viws

A superlight HTTP fileserver with customizable behavior.

[![Build Status](https://travis-ci.org/ViBiOh/viws.svg?branch=master)](https://travis-ci.org/ViBiOh/viws)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/viws)](https://goreportcard.com/report/github.com/ViBiOh/viws)
[![codecov](https://codecov.io/gh/ViBiOh/viws/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/viws)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FViBiOh%2Fviws.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FViBiOh%2Fviws?ref=badge_shield)

## Installation

```bash
go get github.com/ViBiOh/viws/cmd/viws
```

### Light version

Light version (without Opentracing and Prometheus) is also available, for a smaller binary.

```bash
go get github.com/ViBiOh/viws/cmd/viws-light
```

## Features

* Full TLS support
* Opentracing with Jaeger
* Prometheus monitoring
* Read-only container
* Serve static content, with Single Page App handling
* Serve environment variables for easier-config
* Custom 404 page
* Graceful close

## Single Page Application

This mode is useful when you have a router in your javascript framework (e.g. Angular/React/Vue). When a request target a not found file, it returns the index instead of 404. This option also deactivates cache for the index in order to make work the cache-buster for javascript/style files.

e.g.
```bash
curl myWebsite.com/users/vibioh/
=> /index.html
```

Be careful, `-notFound` and `-spa` are incompatible flags. If you set both, you'll get an error.

## Endpoints

* `GET /health`: [healthcheck](#graceful-close) of server
* `GET /version`: value of `VERSION` environment variable
* `GET /env`: values of [specified environments variables](#environment-variables)

## Environment variables

Environment variables are exposed as JSON from a single and easy to remember endpoint: `/env`. You have full control of exposed variables by declaring them on the CLI.

This feature is useful for Single Page Application, you first request `/env` in order to know the `API_URL` or `CONFIGURATION_TOKEN` and then proceed. You reuse the same artifact between `pre-production` and `production`, only variables change.

```bash
API_URL=https://api.vibioh.fr vibioh/viws --env API_URL

> curl http://localhost:1080/env
{"API_URL":"https://api.vibioh.fr"}
```

```js
// index.js

const response = await fetch('/env');
const config = await response.json();
ReactDOM.render(<App />, document.getElementById('root'));
```

## Usage

By default, server is listening on the `1080` port and serve content for GET requests from the `/www/` directory, which have to contains an `index.html`. It assumes that HTTPS is done, somewhere between browser and server (e.g. CloudFlare, ReverseProxy, Traefik, ...) so it sets HSTS flag by default.

```bash
Usage of viws:
  -cert string
        [http] Certificate file
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
  -graceful string
        [http] Graceful close duration (default "35s")
  -headers string
        [viws] Custom headers, tilde separated (e.g. content-language:fr~X-UA-Compatible:test)
  -hsts
        [owasp] Indicate Strict Transport Security (default true)
  -key string
        [http] Key file
  -notFound
        [viws] Graceful 404 page at /404.html (GET request)
  -port int
        [http] Listen port (default 1080)
  -prometheusPath string
        [prometheus] Path for exposing metrics (default "/metrics")
  -push string
        [viws] Paths for HTTP/2 Server Push on index, comma separated
  -spa
        [viws] Indicate Single Page Application mode
  -tracingAgent string
        [tracing] Jaeger Agent (e.g. host:port) (default "jaeger:6831")
  -tracingName string
        [tracing] Service name
  -url string
        [alcotest] URL to check
  -userAgent string
        [alcotest] User-Agent for check (default "Golang alcotest")
```

## Docker

```bash
docker run \
  -d \
  -p 1080:1080/tcp \
  -v "$(pwd):/www/:ro" \
  vibioh/viws
```

We recommend using a Dockerfile to ship your files inside it.

e.g.
```
FROM vibioh/viws

ENV VERSION 1.0.0-1234abcd
COPY dist/ /www/
```

### Light image

Image with tag `:light` is also available.

e.g.
```
FROM vibioh/viws:light

ENV VERSION 1.0.0-1234abcd
COPY dist/ /www/
```

## Compilation

You need Go 1.11+ with go modules enabled in order to compile the project.

```bash
make go
```

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FViBiOh%2Fviws.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FViBiOh%2Fviws?ref=badge_large)
