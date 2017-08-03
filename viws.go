package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/ViBiOh/alcotest/alcotest"
)

const notFoundFilename = `404.html`
const indexFilename = `index.html`

var pngFile = regexp.MustCompile(`.png$`)
var acceptGzip = regexp.MustCompile(`^(?:gzip|\*)(?:;q=(?:1.*?|0\.[1-9][0-9]*))?$`)
var requestsHandler = gzipHandler{owaspHandler{customFileHandler{}}}

var (
	directory = flag.String(`directory`, `/www/`, `Directory to serve`)
	csp       = flag.String(`csp`, `default-src 'self'`, `Content-Security-Policy`)
	notFound  = flag.Bool(`notFound`, false, `Graceful 404 page at /404.html`)
	spa       = flag.Bool(`spa`, false, `Indicate Single Page Application mode`)
	hsts      = flag.Bool(`hsts`, true, `Indicate Strict Transport Security`)
)

var (
	notFoundPath *string
	envKeys      []string
	pushPaths    []string
)

func isFileExist(parts ...string) *string {
	fullPath := path.Join(parts...)
	info, err := os.Stat(fullPath)

	if err != nil {
		log.Print(err)
		return nil
	}

	if info.IsDir() {
		if isFileExist(append(parts, indexFilename)...) == nil {
			return nil
		}
	}

	return &fullPath
}

type gzipMiddleware struct {
	http.ResponseWriter
	gzw *gzip.Writer
}

func (m *gzipMiddleware) WriteHeader(status int) {
	m.ResponseWriter.Header().Add(`Vary`, `Accept-Encoding`)
	m.ResponseWriter.Header().Set(`Content-Encoding`, `gzip`)
	m.ResponseWriter.Header().Del(`Content-Length`)

	if !(http.StatusNotFound == status && *notFound) {
		m.ResponseWriter.WriteHeader(status)
	}
}

func (m *gzipMiddleware) Write(b []byte) (int, error) {
	return m.gzw.Write(b)
}

func (m *gzipMiddleware) Flush() {
	m.gzw.Flush()

	if flusher, ok := m.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (m *gzipMiddleware) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := m.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf(`http.Hijacker not available`)
}

type gzipHandler struct {
	h http.Handler
}

func (handler gzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if acceptEncodingGzip(r) && !pngFile.MatchString(r.URL.Path) {
		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()

		handler.h.ServeHTTP(&gzipMiddleware{w, gzipWriter}, r)
	} else {
		handler.h.ServeHTTP(w, r)
	}
}

func acceptEncodingGzip(r *http.Request) bool {
	header := r.Header.Get(`Accept-Encoding`)

	for _, headerEncoding := range strings.Split(header, `,`) {
		if acceptGzip.MatchString(headerEncoding) {
			return true
		}
	}

	return false
}

type owaspMiddleware struct {
	http.ResponseWriter
	path string
}

func (m *owaspMiddleware) WriteHeader(status int) {
	if status < http.StatusBadRequest {
		m.Header().Add(`Content-Security-Policy`, *csp)
		m.Header().Add(`Referrer-Policy`, `strict-origin-when-cross-origin`)
		m.Header().Add(`X-Frame-Options`, `deny`)
		m.Header().Add(`X-Content-Type-Options`, `nosniff`)
		m.Header().Add(`X-XSS-Protection`, `1; mode=block`)
		m.Header().Add(`X-Permitted-Cross-Domain-Policies`, `none`)
	}

	if *hsts {
		m.Header().Add(`Strict-Transport-Security`, `max-age=5184000`)
	}

	if status == http.StatusOK || status == http.StatusMovedPermanently {
		if m.path == `/` {
			m.Header().Add(`Cache-Control`, `no-cache`)
		} else {
			m.Header().Add(`Cache-Control`, `max-age=864000`)
		}
	}

	m.ResponseWriter.WriteHeader(status)
}

type owaspHandler struct {
	h http.Handler
}

func (handler owaspHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.h.ServeHTTP(&owaspMiddleware{w, r.URL.Path}, r)
}

type customFileHandler struct {
}

func (handler customFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	if filePath := isFileExist(*directory, r.URL.Path); filePath != nil {
		http.ServeFile(w, r, *filePath)
	} else if *notFound {
		w.WriteHeader(http.StatusNotFound)
		http.ServeFile(w, r, *notFoundPath)
	} else if *spa {
		http.ServeFile(w, r, *directory)
	} else {
		http.Error(w, `404 page not found: `+r.URL.Path, http.StatusNotFound)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func envHandler(w http.ResponseWriter, r *http.Request) {
	env := make(map[string]string)

	for _, key := range envKeys {
		if value := os.Getenv(key); value != `` {
			env[key] = value
		}
	}

	if objJSON, err := json.Marshal(env); err == nil {
		w.Header().Set(`Content-Type`, `application/json`)
		w.Header().Set(`Cache-Control`, `no-cache`)
		w.Header().Set(`Access-Control-Allow-Origin`, `*`)
		w.Write(objJSON)
	} else {
		http.Error(w, `Error while marshalling JSON response`, http.StatusInternalServerError)
	}
}

func viwsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == `/health` {
		healthHandler(w, r)
	} else if r.URL.Path == `/env` {
		envHandler(w, r)
	} else {
		requestsHandler.ServeHTTP(w, r)
	}
}

func handleGracefulClose(server *http.Server) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	<-signals
	log.Print(`SIGTERM received`)

	if server != nil {
		log.Print(`Shutting down http server`)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Print(err)
		}
	}
}

func main() {
	url := flag.String(`c`, ``, `URL to healthcheck (check and exit)`)
	port := flag.String(`port`, `1080`, `Listening port`)
	keys := flag.String(`env`, ``, `Environments key variables to expose, comma separated`)
	push := flag.String(`push`, ``, `Paths for HTTP/2 Server Push, comma separated`)
	https := flag.Bool(`https`, false, `Serve TLS content`)
	cert := flag.String(`cert`, `cert.pem`, `Certificate filename for HTTPS`)
	key := flag.String(`key`, `key.pem`, `Key filename for HTTPS`)
	flag.Parse()

	if *url != `` {
		alcotest.Do(url)
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	if isFileExist(*directory) == nil {
		log.Fatalf(`Directory %s is unreachable`, *directory)
	}

	log.Printf(`Starting server on port %s`, *port)
	log.Printf(`Serving file from %s`, *directory)
	log.Printf(`Content-Security-Policy: %s`, *csp)

	if *spa {
		log.Print(`Working in SPA mode`)
	}

	if *keys != `` {
		envKeys = strings.Split(*keys, `,`)
	}

	if *push != `` {
		pushPaths = strings.Split(*push, `,`)

		if !*https {
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
		Handler: http.HandlerFunc(viwsHandler),
	}

	if *https {
		go server.ListenAndServeTLS(*cert, *key)
	} else {
		go server.ListenAndServe()
	}
	handleGracefulClose(server)
}
