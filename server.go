package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
)

const tenDaysOfCaching = `864000`
const contentSecurityPolicy = `default-src 'self' 'unsafe-inline' `
const notFoundFilename = `404.html`
const indexFilename = `index.html`

var domain string
var notFound bool
var spa bool
var hsts bool

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

type owaspMiddleware struct {
	http.ResponseWriter
	path string
}

func (m *owaspMiddleware) WriteHeader(status int) {
	if status < http.StatusBadRequest {
		m.Header().Add(`Content-Security-Policy`, contentSecurityPolicy+domain)
		m.Header().Add(`X-Frame-Options`, `deny`)
		m.Header().Add(`X-Content-Type-Options`, `nosniff`)
		m.Header().Add(`X-XSS-Protection`, `1; mode=block`)
		m.Header().Add(`X-Permitted-Cross-Domain-Policies`, `none`)
	}

	if hsts {
		m.Header().Add(`Strict-Transport-Security`, `max-age=`+tenDaysOfCaching)
	}

	if status == http.StatusOK || status == http.StatusMovedPermanently {
		if spa && m.path == `/` {
			m.Header().Add(`Cache-Control`, `no-cache`)
		} else {
			m.Header().Add(`Cache-Control`, `max-age=`+tenDaysOfCaching)
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
		http.Error(w, `Method Not Allowed`, http.StatusMethodNotAllowed)
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

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	port := flag.String(`port`, `1080`, `Listening port`)
	directory := flag.String(`directory`, `/www/`, `Directory to serve`)
	flag.BoolVar(&hsts, `hsts`, true, `Indicate Strict Transport Security`)
	flag.BoolVar(&spa, `spa`, false, `Indicate Single Page Application mode`)
	flag.BoolVar(&notFound, `notFound`, false, `Graceful 404 page at /404.html`)
	flag.StringVar(&domain, `domain`, ``, `Domains names for Content-Security-Policy appended to "default-src 'self' 'unsafe-inline'"`)
	flag.Parse()

	if isFileExist(*directory) == nil {
		log.Fatal(`Directory ` + *directory + ` is unreachable.`)
	}

	log.Println(`Starting server on port ` + *port)
	log.Println(`Serving file from ` + *directory)
	log.Println(`Content-Security-Policy: `, contentSecurityPolicy+domain)

	if spa {
		log.Println(`Working in SPA mode`)
	}

	var notFoundPath *string

	if notFound {
		if notFoundPath = isFileExist(*directory, notFoundFilename); notFoundPath == nil {
			log.Println(*directory + notFoundFilename + ` is unreachable. Flag ignored.`)
			notFound = false
		} else {
			log.Println(`404 will be ` + *notFoundPath)
		}
	}

	http.Handle(`/`, owaspHandler{customFileHandler{directory, notFoundPath}})

	log.Fatal(http.ListenAndServe(`:`+*port, nil))
}
