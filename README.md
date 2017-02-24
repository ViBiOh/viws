# go-http

A superlight HTTP fileserver with customizable behavior.

[![Build Status](https://travis-ci.org/ViBiOh/go-http.svg?branch=master)](https://travis-ci.org/ViBiOh/go-http) [![](https://images.microbadger.com/badges/image/vibioh/http.svg)](https://microbadger.com/images/vibioh/http "Get your own image badge on microbadger.com")

## Compilation

You need Go in order to compile the project.

```
make
```

## Running

## Docker

`docker run -d -p 1080:1080 -v /var/www/html/:/www/ vibioh/http`

We recommend using a Dockerfile to ship your files inside it.

e.g.
```
FROM vibioh/http

COPY dist/ /www/
```

## Usage

By default, server is listening on the `1080` port and serve content for GET requests from the `/www/` directory. It assumes that HTTPS is done somewhere between browser and server (e.g. CloudFlare, ReverseProxy, Traefik, ...) so it sets HSTS flag.

```
Usage of ./server:
  -directory string
      Directory to serve (default "/www/")
  -csp string
      Content-Security-Policy (default "default-src 'self'")
  -hsts
      Indicate Strict Transport Security (default true)
  -notFound
      Graceful 404 page at /404.html
  -port string
      Listening port (default "1080")
  -spa
      Indicate Single Page Application mode
```

## Single Page Application

This mode is useful when you have a router in your javascript framework (e.g. Angular/React). When a request target a not found file, it returns the index instead of 404.

e.g.
```
curl myWebsite.com/users/vibioh/
=> /index.html
```

Be careful, `-notFound` and `-spa` are incompatible flags. If you set both, the `-notFound` has priority over the `-spa`.
