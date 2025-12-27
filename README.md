# viws

A superlight HTTP fileserver with customizable behavior.

[![Build](https://github.com/ViBiOh/viws/workflows/Build/badge.svg)](https://github.com/ViBiOh/viws/actions)

## Installation

```bash
go install github.com/ViBiOh/viws/cmd/viws@latest
```

### Light version

Light version (without GZIP and Open Telemetry) is also available, for a smaller binary.

```bash
go install github.com/ViBiOh/viws/cmd/viws-light@latest
```

## Features

- Full TLS support
- GZIP Compression
- Open Telemetry observability
- Read-only container
- Serve static content, with Single Page App handling
- Serve environment variables for easier-config
- Configurable logger with JSON support

## Single Page Application

This mode is useful when you have a router in your javascript framework (e.g. Angular/React/Vue). When a request target a not found file, it returns the index instead of 404. This option also deactivates cache for the index in order to make work the cache-buster for javascript/style files.

```bash
curl myWebsite.com/users/vibioh/
=> /index.html
```

## Endpoints

- `GET /health`: healthcheck of server, always respond [`okStatus (default 204)`](#usage)
- `GET /ready`: checks external dependencies availability and then respond [`okStatus (default 204)`](#usage) or `503` during [`graceDuration`](#usage) when close signal is received
- `GET /version`: value of `VERSION` environment variable
- `GET /env`: values of [specified environments variables](#environment-variables)

## Environment variables

Environment variables are exposed as JSON from a single and easy to remember endpoint: `/env`. You have full control of exposed variables by declaring them on the CLI.

This feature is useful for Single Page Application, you first request `/env` in order to know the `API_URL` or `CONFIGURATION_TOKEN` and then proceed. You reuse the same artifact between `pre-production` and `production`, only variables change, in [respect of 12factor app](https://12factor.net/config)

### Configuration example

```bash
API_URL=https://api.vibioh.fr vibioh/viws --env API_URL

> curl http://127.0.0.1:1080/env
{"API_URL":"https://api.vibioh.fr"}
```

### Usage in SPA

```js
// index.js

const response = await fetch("/env");
const config = await response.json();
ReactDOM.render(<App config={config} />, document.getElementById("root"));
```

## Usage

By default, server is listening on the `1080` port and serve content for GET requests from the `/www/` directory. It assumes that HTTPS is done, somewhere between browser and server (e.g. CloudFlare, ReverseProxy, Traefik, ...) so it sets HSTS flag by default.

The application can be configured by passing CLI args described below or their equivalent as environment variable. CLI values take precedence over environments variables.

Be careful when using the CLI values, if someone list the processes on the system, they will appear in plain-text. Pass secrets by environment variables: it's less easily visible.

```bash
Usage of viws:
  --address           string        [server] Listen address ${VIWS_ADDRESS}
  --cert              string        [server] Certificate file ${VIWS_CERT}
  --corsCredentials                 [cors] Access-Control-Allow-Credentials ${VIWS_CORS_CREDENTIALS} (default false)
  --corsExpose        string        [cors] Access-Control-Expose-Headers ${VIWS_CORS_EXPOSE}
  --corsHeaders       string        [cors] Access-Control-Allow-Headers ${VIWS_CORS_HEADERS} (default "Content-Type")
  --corsMethods       string        [cors] Access-Control-Allow-Methods ${VIWS_CORS_METHODS} (default "GET")
  --corsOrigin        string        [cors] Access-Control-Allow-Origin ${VIWS_CORS_ORIGIN} (default "*")
  --csp               string        [owasp] Content-Security-Policy ${VIWS_CSP} (default "default-src 'self'; base-uri 'self'")
  --directory         string        [viws] Directory to serve ${VIWS_DIRECTORY} (default "/www/")
  --env               string slice  [env] Environments key variable to expose ${VIWS_ENV}, as a string slice, environment variable separated by ","
  --frameOptions      string        [owasp] X-Frame-Options ${VIWS_FRAME_OPTIONS} (default "deny")
  --graceDuration     duration      [http] Grace duration when signal received ${VIWS_GRACE_DURATION} (default 30s)
  --gzip                            [gzip] Enable gzip compression ${VIWS_GZIP} (default true)
  --header            string slice  [viws] Custom header e.g. content-language:fr ${VIWS_HEADER}, as a string slice, environment variable separated by ","
  --hsts                            [owasp] Indicate Strict Transport Security ${VIWS_HSTS} (default true)
  --idleTimeout       duration      [server] Idle Timeout ${VIWS_IDLE_TIMEOUT} (default 2m0s)
  --key               string        [server] Key file ${VIWS_KEY}
  --loggerJson                      [logger] Log format as JSON ${VIWS_LOGGER_JSON} (default false)
  --loggerLevel       string        [logger] Logger level ${VIWS_LOGGER_LEVEL} (default "INFO")
  --loggerLevelKey    string        [logger] Key for level in JSON ${VIWS_LOGGER_LEVEL_KEY} (default "level")
  --loggerMessageKey  string        [logger] Key for message in JSON ${VIWS_LOGGER_MESSAGE_KEY} (default "msg")
  --loggerTimeKey     string        [logger] Key for timestamp in JSON ${VIWS_LOGGER_TIME_KEY} (default "time")
  --name              string        [server] Name ${VIWS_NAME} (default "http")
  --okStatus          int           [http] Healthy HTTP Status code ${VIWS_OK_STATUS} (default 204)
  --port              uint          [server] Listen port (0 to disable) ${VIWS_PORT} (default 1080)
  --pprofAgent        string        [pprof] URL of the Datadog Trace Agent (e.g. http://datadog.observability:8126) ${VIWS_PPROF_AGENT}
  --pprofPort         int           [pprof] Port of the HTTP server (0 to disable) ${VIWS_PPROF_PORT} (default 0)
  --readTimeout       duration      [server] Read Timeout ${VIWS_READ_TIMEOUT} (default 5s)
  --shutdownTimeout   duration      [server] Shutdown Timeout ${VIWS_SHUTDOWN_TIMEOUT} (default 10s)
  --spa                             [viws] Indicate Single Page Application mode ${VIWS_SPA} (default false)
  --telemetryRate     string        [telemetry] OpenTelemetry sample rate, 'always', 'never' or a float value ${VIWS_TELEMETRY_RATE} (default "always")
  --telemetryURL      string        [telemetry] OpenTelemetry gRPC endpoint (e.g. otel-exporter:4317) ${VIWS_TELEMETRY_URL}
  --telemetryUint64                 [telemetry] Change OpenTelemetry Trace ID format to an unsigned int 64 ${VIWS_TELEMETRY_UINT64} (default true)
  --url               string        [alcotest] URL to check ${VIWS_URL}
  --userAgent         string        [alcotest] User-Agent for check ${VIWS_USER_AGENT} (default "Alcotest")
  --writeTimeout      duration      [server] Write Timeout ${VIWS_WRITE_TIMEOUT} (default 10s)
```

## Docker

```bash
docker run -d --name website \
  -p 1080:1080/tcp \
  -v "$(pwd):/www/:ro" \
  vibioh/viws
```

We recommend using a Dockerfile to ship your files inside it.

e.g.

```
FROM rg.fr-par.scw.cloud/vibioh/viws

ENV VERSION 1.2.3-1234abcd
COPY dist/ /www/
```

### Light image

Image with tag `:light` is also available.

e.g.

```
FROM rg.fr-par.scw.cloud/vibioh/viws:light

ENV VERSION 1.0.0-1234abcd
COPY dist/ /www/
```

## Compilation

```bash
make go
```
