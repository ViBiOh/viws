# viws

A superlight HTTP fileserver with customizable behavior.

[![Build Status](https://travis-ci.org/ViBiOh/viws.svg?branch=master)](https://travis-ci.org/ViBiOh/viws)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/viws)](https://goreportcard.com/report/github.com/ViBiOh/viws)

## Installation

```
go get -u github.com/ViBiOh/viws
```

## Usage

By default, server is listening on the `1080` port and serve content for GET requests from the `/www/` directory, which have to contains an `index.html`. It assumes that HTTPS is done somewhere between browser and server (e.g. CloudFlare, ReverseProxy, Traefik, ...) so it sets HSTS flag.

```
Usage of viws:
  -c string
    	URL to healthcheck (check and exit)
  -cert string
    	Certificate filename for TLS (default "cert.pem")
  -csp string
    	Content-Security-Policy (default "default-src 'self'")
  -directory string
    	Directory to serve (default "/www/")
  -env string
    	Environments key variables to expose, comma separated
  -hsts
    	Indicate Strict Transport Security (default true)
  -key string
    	Key filename for TLS (default "key.pem")
  -notFound
    	Graceful 404 page at /404.html
  -port string
    	Listening port (default "1080")
  -push string
    	Paths for HTTP/2 Server Push, comma separated
  -spa
    	Indicate Single Page Application mode
  -tls
    	Serve TLS content
```

## Single Page Application

This mode is useful when you have a router in your javascript framework (e.g. Angular/React). When a request target a not found file, it returns the index instead of 404. This option also deactivates cache for the index in order to make work the cache-buster for javascript/style files.

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

You need Go in order to compile the project.

```
make
```
