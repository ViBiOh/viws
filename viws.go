package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/owasp"
	"github.com/ViBiOh/httputils/prometheus"
	"github.com/ViBiOh/httputils/rate"
	"github.com/ViBiOh/viws/env"
	"github.com/ViBiOh/viws/viws"
)

var (
	requestsHandler http.Handler
	envHandler      http.Handler
	apiHandler      http.Handler
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
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

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
	tls := flag.Bool(`tls`, false, `Serve TLS content`)
	alcotestConfig := alcotest.Flags(``)
	certConfig := cert.Flags(`tls`)
	prometheusConfig := prometheus.Flags(`prometheus`)
	rateConfig := rate.Flags(`rate`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)

	viwsConfig := viws.Flags(``)
	envConfig := env.Flags(``)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	viwsApp, err := viws.NewApp(viwsConfig, *tls)
	if err != nil {
		log.Fatalf(`Error while instanciating viws: %v`, err)
	}

	requestsHandler = viwsApp.ServerPushHandler(owasp.Handler(owaspConfig, viwsApp.FileHandler()))
	envHandler = owasp.Handler(owaspConfig, cors.Handler(corsConfig, env.Handler(envConfig)))
	apiHandler = prometheus.Handler(prometheusConfig, rate.Handler(rateConfig, gziphandler.GzipHandler(handler())))

	log.Printf(`Starting server on port %s`, *port)
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
