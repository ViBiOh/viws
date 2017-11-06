package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/owasp"
	"github.com/ViBiOh/httputils/prometheus"
	"github.com/ViBiOh/httputils/rate"
	"github.com/ViBiOh/httputils/writer"
	"github.com/ViBiOh/viws/env"
)

const notFoundFilename = `404.html`
const indexFilename = `index.html`

var requestsHandler http.Handler
var envHandler http.Handler
var apiHandler http.Handler

var (
	directory = flag.String(`directory`, `/www/`, `Directory to serve`)
	notFound  = flag.Bool(`notFound`, false, `Graceful 404 page at /404.html`)
	spa       = flag.Bool(`spa`, false, `Indicate Single Page Application mode`)
)

var (
	notFoundPath *string
	pushPaths    []string
)

func isFileExist(parts ...string) *string {
	fullPath := path.Join(parts...)
	info, err := os.Stat(fullPath)

	if err != nil {
		return nil
	}

	if info.IsDir() {
		if isFileExist(append(parts, indexFilename)...) == nil {
			return nil
		}
	}

	return &fullPath
}

func serverPushHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path == `/` && len(pushPaths) > 0 {
			if pusher, ok := w.(http.Pusher); ok {
				for _, path := range pushPaths {
					if err := pusher.Push(path, nil); err != nil {
						log.Printf(`Failed to push %s: %v`, path, err)
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func fileHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fakeWriter := writer.ResponseWriter{}
		http.ServeFile(&fakeWriter, r, *directory+r.URL.Path)

		if fakeWriter.Status() == http.StatusNotFound && (*notFound || *spa) {
			if *notFound {
				fakeWriter = writer.ResponseWriter{}
				http.ServeFile(&fakeWriter, r, *notFoundPath)
				fakeWriter.SetStatus(http.StatusNotFound)
			} else if *spa {
				fakeWriter = writer.ResponseWriter{}
				http.ServeFile(&fakeWriter, r, *directory)
			}
		}

		for k, v := range fakeWriter.Header() {
			w.Header()[k] = v
		}

		w.WriteHeader(fakeWriter.Status())
		if fakeWriter.Content() != nil {
			w.Write(fakeWriter.Content().Bytes())
		}
	})
}

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
	url := flag.String(`c`, ``, `URL to healthcheck (check and exit)`)
	port := flag.String(`port`, `1080`, `Listening port`)
	push := flag.String(`push`, ``, `Paths for HTTP/2 Server Push, comma separated`)
	tls := flag.Bool(`tls`, false, `Serve TLS content`)
	prometheusConfig := prometheus.Flags(`prometheus`)
	rateConfig := rate.Flags(`rate`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)
	flag.Parse()

	if *url != `` {
		alcotest.Do(url)
		return
	}

	if err := env.Init(); err != nil {
		log.Fatalf(`Error while initializing env: %v`, err)
	}

	if isFileExist(*directory) == nil {
		log.Fatalf(`Directory %s is unreachable`, *directory)
	}

	log.Printf(`Starting server on port %s`, *port)
	log.Printf(`Serving file from %s`, *directory)

	if *spa {
		log.Print(`Working in SPA mode`)
	}

	if *push != `` {
		pushPaths = strings.Split(*push, `,`)

		if !*tls {
			log.Print(`⚠ HTTP/2 Server Push works only when TLS in enabled ⚠`)
		}
	}

	if *notFound {
		if *spa {
			log.Print(`⚠ -notFound and -spa are both set. SPA flag is ignored ⚠`)
		}

		if notFoundPath = isFileExist(*directory, notFoundFilename); notFoundPath == nil {
			log.Printf(`%s%s is unreachable. Not found flag ignored.`, *directory, notFoundFilename)
			*notFound = false
		} else {
			log.Printf(`404 will be %s`, *notFoundPath)
		}
	}

	requestsHandler = serverPushHandler(owasp.Handler(owaspConfig, fileHandler()))
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
			serveError <- cert.ListenAndServeTLS(server)
		} else {
			serveError <- server.ListenAndServe()
		}
	}()

	httputils.ServerGracefulClose(server, serveError, nil)
}
