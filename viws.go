package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/owasp"
	"github.com/ViBiOh/httputils/prometheus"
	"github.com/ViBiOh/httputils/rate"
	"github.com/ViBiOh/viws/env"
	"github.com/ViBiOh/viws/utils"
	"github.com/ViBiOh/viws/viws"
)

const notFoundFilename = `404.html`

var requestsHandler http.Handler
var envHandler http.Handler
var apiHandler http.Handler

var (
	directory = flag.String(`directory`, `/www/`, `Directory to serve`)
	notFound  = flag.Bool(`notFound`, false, `Graceful 404 page at /404.html`)
	spa       = flag.Bool(`spa`, false, `Indicate Single Page Application mode`)
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == `/health` {
			healthHandler(w, r)
		} else if r.URL.Path == `/env` {
			envHandler.ServeHTTP(w, r)
		} else {
			requestsHandler.ServeHTTP(w, r)
		}
	})
}

func main() {
	port := flag.String(`port`, `1080`, `Listening port`)
	push := flag.String(`push`, ``, `Paths for HTTP/2 Server Push, comma separated`)
	tls := flag.Bool(`tls`, false, `Serve TLS content`)
	alcotestConfig := alcotest.Flags(``)
	certConfig := cert.Flags(`tls`)
	prometheusConfig := prometheus.Flags(`prometheus`)
	rateConfig := rate.Flags(`rate`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)
	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	if err := env.Init(); err != nil {
		log.Fatalf(`Error while initializing env: %v`, err)
	}

	if utils.IsFileExist(*directory) == nil {
		log.Fatalf(`Directory %s is unreachable or does not contains index`, *directory)
	}

	log.Printf(`Starting server on port %s`, *port)
	log.Printf(`Serving file from %s`, *directory)

	if *spa {
		log.Print(`Working in SPA mode`)
	}

	if *push != `` {
		if !*tls {
			log.Print(`⚠ HTTP/2 Server Push works only when TLS in enabled ⚠`)
		}
	}

	var notFoundPath *string
	if *notFound {
		if *spa {
			log.Print(`⚠ -notFound and -spa are both set. SPA flag is ignored ⚠`)
		}

		if notFoundPath = utils.IsFileExist(*directory, notFoundFilename); notFoundPath == nil {
			log.Printf(`%s%s is unreachable. Not found flag ignored.`, *directory, notFoundFilename)
			*notFound = false
		} else {
			log.Printf(`404 will be %s`, *notFoundPath)
		}
	}

	requestsHandler = viws.ServerPushHandler(owasp.Handler(owaspConfig, viws.FileHandler(*directory, *spa, notFoundPath)), strings.Split(*push, `,`))
	envHandler = owasp.Handler(owaspConfig, cors.Handler(corsConfig, env.Handler()))
	apiHandler = prometheus.Handler(prometheusConfig, rate.Handler(rateConfig, gziphandler.GzipHandler(handler())))

	server := &http.Server{
		Addr:    `:` + *port,
		Handler: apiHandler,
	}

	var serveError = make(chan error)
	go func() {
		defer close(serveError)
		if *tls {
			log.Print(`Listening with TLS enabled`)
			serveError <- cert.ListenAndServeTLS(certConfig, server)
		} else {
			serveError <- server.ListenAndServe()
		}
	}()

	httputils.ServerGracefulClose(server, serveError, nil)
}
