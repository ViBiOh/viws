# go-http

A superlight HTTP FileServer with custom behavior.

[![Build Status](https://travis-ci.org/ViBiOh/go-http.svg?branch=master)](https://travis-ci.org/ViBiOh/go-http) [![](https://images.microbadger.com/badges/image/vibioh/http.svg)](https://microbadger.com/images/vibioh/http "Get your own image badge on microbadger.com")

## Compilation

You need Go in order to compile the project.

```
make
```

## Usage

```
./server
```

By default, server is listening GET requests on the `1080` port and serve content from the `/www/` directory. It assumes that HTTPS is done somewhere between browser and server (e.g. CloudFlare) so it sets HSTS flag.

All OWASP security headers are defined for your browser. You can override `Content-Security-Policy` with arg `-domain`. String will be appended at the end of the `default-src 'self' 'unsafe-inline' ` predefined policy.

If you want to redirect to the `/404.html` page when not found, set the `-notFound` flag.

If you work with a Single Page Application (Angular / React), it can be interesting to redirect all not found requests to index (and let the router of the framework handle it). This particular mode can be set with `-spa` flag.

Be careful, `-notFound` and `-spa` are incompatible flags. If you set both, the `-notFound` has priority over the `-spa`.
