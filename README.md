# viws

A superlight HTTP fileserver with customizable behavior.

[![Build Status](https://travis-ci.org/ViBiOh/viws.svg?branch=master)](https://travis-ci.org/ViBiOh/viws)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/viws)](https://goreportcard.com/report/github.com/ViBiOh/viws)

## Installation

```
go get -u github.com/ViBiOh/viws
```

## Usage

By default, server is listening on the `1080` port and serve content for GET requests from the `/www/` directory, which have to contains an `index.html`. It assumes that HTTPS is done, somewhere between browser and server (e.g. CloudFlare, ReverseProxy, Traefik, self-signed, ...) so it sets HSTS flag by default, security matters.

```
Usage of viws:
  -c string
        URL to healthcheck (check and exit)
  -corsHeaders string
        Access-Control-Allow-Headers (default "Content-Type")
  -corsMethods string
        Access-Control-Allow-Methods (default "GET")
  -corsOrigin string
        Access-Control-Allow-Origin (default "*")
  -csp string
        Content-Security-Policy (default "default-src 'self'")
  -directory string
        Directory to serve (default "/www/")
  -env string
        Environments key variables to expose, comma separated
  -hsts
        Indicate Strict Transport Security (default true)
  -notFound
        Graceful 404 page at /404.html
  -port string
        Listening port (default "1080")
  -prometheusMetricsHost string
        Prometheus - Allowed hostname to call metrics endpoint (default "localhost")
  -prometheusMetricsPath string
        Prometheus - Metrics endpoint path (default "/metrics")
  -push string
        Paths for HTTP/2 Server Push, comma separated
  -rateCount int
        Rate IP limit (default 5000)
  -spa
        Indicate Single Page Application mode
  -tls
        Serve TLS content
  -tlscert string
        TLS PEM Certificate file
  -tlshosts string
        TLS Self-signed certificate hosts, comma separated (default "localhost")
  -tlskey string
        TLS PEM Key file
```

## Single Page Application

This mode is useful when you have a router in your javascript framework (e.g. Angular/React/Vue). When a request target a not found file, it returns the index instead of 404. This option also deactivates cache for the index in order to make work the cache-buster for javascript/style files.

e.g.
```
curl myWebsite.com/users/vibioh/
=> /index.html
```

Be careful, `-notFound` and `-spa` are incompatible flags. If you set both, the `-notFound` has priority over the `-spa`.


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

```
make
```
