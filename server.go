package main

import (
	"bufio"
	"compress/gzip"
	"context"
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
)

const notFoundFilename = `404.html`
const indexFilename = `index.html`

var pngFile = regexp.MustCompile(`.png$`)
var acceptGzip = regexp.MustCompile(`^(?:gzip|\*)(?:;q=(?:1.*?|0\.[1-9][0-9]*))?$`)

var csp string
var notFound bool
var spa bool
var hsts bool

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

	if !(http.StatusNotFound == status && notFound) {
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
		m.Header().Add(`Content-Security-Policy`, csp)
		m.Header().Add(`Referrer-Policy`, `strict-origin-when-cross-origin`)
		m.Header().Add(`X-Frame-Options`, `deny`)
		m.Header().Add(`X-Content-Type-Options`, `nosniff`)
		m.Header().Add(`X-XSS-Protection`, `1; mode=block`)
		m.Header().Add(`X-Permitted-Cross-Domain-Policies`, `none`)
	}

	if hsts {
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
	root         *string
	notFoundPath *string
}

func (handler customFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else if filePath := isFileExist(*handler.root, r.URL.Path); filePath != nil {
		http.ServeFile(w, r, *filePath)
	} else if notFound {
		w.WriteHeader(http.StatusNotFound)
		http.ServeFile(w, r, *handler.notFoundPath)
	} else if spa {
		http.ServeFile(w, r, *handler.root)
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

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	port := flag.String(`port`, `1080`, `Listening port`)
	directory := flag.String(`directory`, `/www/`, `Directory to serve`)
	flag.BoolVar(&hsts, `hsts`, true, `Indicate Strict Transport Security`)
	flag.BoolVar(&spa, `spa`, false, `Indicate Single Page Application mode`)
	flag.BoolVar(&notFound, `notFound`, false, `Graceful 404 page at /404.html`)
	flag.StringVar(&csp, `csp`, `default-src 'self'`, `Content-Security-Policy`)
	flag.Parse()

	if isFileExist(*directory) == nil {
		log.Fatal(`Directory ` + *directory + ` is unreachable.`)
	}

	log.Println(`Starting server on port`, *port)
	log.Println(`Serving file from`, *directory)
	log.Println(`Content-Security-Policy:`, csp)

	if spa {
		log.Println(`Working in SPA mode`)
	}

	var notFoundPath *string

	if notFound {
		if notFoundPath = isFileExist(*directory, notFoundFilename); notFoundPath == nil {
			log.Println(*directory + notFoundFilename + ` is unreachable. Flag ignored.`)
			notFound = false
		} else {
			log.Println(`404 will be`, *notFoundPath)
		}
	}

	http.HandleFunc(`/health`, healthHandler)
	http.Handle(`/`, gzipHandler{owaspHandler{customFileHandler{directory, notFoundPath}}})
	
	signals := make(chan os.Signal, 1)		
	signal.Notify(signals, syscall.SIGTERM)		

	go func() {
		<-signals
		if err := http.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	log.Fatal(http.ListenAndServe(`:`+*port, nil))
}
