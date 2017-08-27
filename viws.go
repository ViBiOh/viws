package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/gzip"
	"github.com/ViBiOh/httputils/owasp"
	"github.com/ViBiOh/httputils/prometheus"
	"github.com/ViBiOh/viws/env"
)

const notFoundFilename = `404.html`
const indexFilename = `index.html`

var requestsHandler = serverPushHandler{gzip.Handler{Handler: owasp.Handler{Handler: customFileHandler{}}}}
var envHandler = owasp.Handler{Handler: cors.Handler{Handler: env.Handler{}}}

type fakeResponseWriter struct {
	status  int
	header  http.Header
	content *bytes.Buffer
}

func (w *fakeResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}

	return w.header
}

func (w *fakeResponseWriter) Write(content []byte) (int, error) {
	if w.content == nil {
		w.content = bytes.NewBuffer([]byte{})
	}

	return w.content.Write(content)
}

func (w *fakeResponseWriter) WriteHeader(status int) {
	w.status = status
}

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

type serverPushHandler struct {
	http.Handler
}

func (handler serverPushHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	handler.Handler.ServeHTTP(w, r)
}

type customFileHandler struct {
}

func (handler customFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fakeWriter := fakeResponseWriter{}
	http.ServeFile(&fakeWriter, r, *directory+r.URL.Path)

	if fakeWriter.status == http.StatusNotFound && (*notFound || *spa) {
		if *notFound {
			w.WriteHeader(http.StatusNotFound)
			http.ServeFile(w, r, *notFoundPath)
		} else if *spa {
			http.ServeFile(w, r, *directory)
		}
	} else {
		for k, v := range fakeWriter.header {
			w.Header()[k] = v
		}
		w.WriteHeader(fakeWriter.status)
		if fakeWriter.content != nil {
			w.Write(fakeWriter.content.Bytes())
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func viwsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == `/health` {
		healthHandler(w, r)
	} else if r.URL.Path == `/env` {
		envHandler.ServeHTTP(w, r)
	} else {
		requestsHandler.ServeHTTP(w, r)
	}
}

func main() {
	url := flag.String(`c`, ``, `URL to healthcheck (check and exit)`)
	port := flag.String(`port`, `1080`, `Listening port`)
	push := flag.String(`push`, ``, `Paths for HTTP/2 Server Push, comma separated`)
	tls := flag.Bool(`tls`, false, `Serve TLS content`)
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
			log.Print(`HTTP/2 Server push works only when TLS in enabled`)
		}
	}

	if *notFound {
		if notFoundPath = isFileExist(*directory, notFoundFilename); notFoundPath == nil {
			log.Printf(`%s%s is unreachable. Not found flag ignored.`, *directory, notFoundFilename)
			*notFound = false
		} else {
			log.Printf(`404 will be %s`, *notFoundPath)
		}
	}

	server := &http.Server{
		Addr:    `:` + *port,
		Handler: prometheus.NewPrometheusHandler(`http`, http.HandlerFunc(viwsHandler)),
	}

	if *tls {
		log.Printf(`Listening with TLS enabled`)
		go log.Print(cert.ListenAndServeTLS(server))
	} else {
		go log.Print(server.ListenAndServe())
	}

	httputils.ServerGracefulClose(server, nil)
}
