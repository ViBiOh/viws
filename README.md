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
      [health] URL to check
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
      [owasp] Content-Security-Policy (default "default-src 'self'")
  -directory string
      [viws] Directory to serve (default "/www/")
  -env string
      [env] Environments key variables to expose, comma separated
  -frameOptions string
      [owasp] X-Frame-Options (default "deny")
  -hsts
      [owasp] Indicate Strict Transport Security (default true)
  -notFound
      [viws] Graceful 404 page at /404.html
  -port string
      Listen port (default "1080")
  -push string
      [viws] Paths for HTTP/2 Server Push on index, comma separated
  -spa
      [viws] Indicate Single Page Application mode
  -tls
      Serve TLS content
  -tlsCert string
      [tls] PEM Certificate file
  -tlsHosts string
      [tls] Self-signed certificate hosts, comma separated (default "localhost")
  -tlsKey string
      [tls] PEM Key file
```

## Single Page Application

This mode is useful when you have a router in your javascript framework (e.g. Angular/React/Vue). When a request target a not found file, it returns the index instead of 404. This option also deactivates cache for the index in order to make work the cache-buster for javascript/style files.

e.g.
```
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

```
make go
```
