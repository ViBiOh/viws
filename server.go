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
const NOT_FOUND_FILENAME = `404.html`
const INDEX_FILENAME = `index.html`

var domain string

func isFileExist(parts ...string) *string {
	fullPath := path.Join(parts...)
	info, err := os.Stat(fullPath)

	if err != nil {
		return nil
	}

	if info.IsDir() {
		if isFileExist(append(parts, INDEX_FILENAME)...) == nil {
			return nil
		}
	}

	return &fullPath
}

type OwaspMiddleware struct {
	http.ResponseWriter
}

func (m *OwaspMiddleware) WriteHeader(status int) {
	if status < http.StatusBadRequest {
		m.Header().Add(`Content-Security-Policy`, contentSecurityPolicy+domain)
		m.Header().Add(`X-Frame-Options`, `deny`)
		m.Header().Add(`X-Content-Type-Options`, `nosniff`)
		m.Header().Add(`X-XSS-Protection`, `1; mode=block`)
	}

	if status == http.StatusOK || status == http.StatusMovedPermanently {
		m.Header().Add(`Cache-Control`, `max-age=`+tenDaysOfCaching)
	}

	m.ResponseWriter.WriteHeader(status)
}

type OwaspHandler struct {
	h http.Handler
}

func (handler OwaspHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.h.ServeHTTP(&OwaspMiddleware{ResponseWriter: w}, r)
}

type CustomFileHandler struct {
	root         *string
	spa          bool
	notFound     bool
	notFoundPath *string
}

func (handler CustomFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if filePath := isFileExist(*handler.root, r.URL.Path); filePath != nil {
		http.ServeFile(w, r, *filePath)
	} else if handler.notFound {
		http.ServeFile(w, r, *handler.notFoundPath)
	} else if handler.spa {
		http.ServeFile(w, r, *handler.root)
	} else {
		http.Error(w, `404 page not found: `+r.URL.Path, 404)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	port := flag.String(`port`, `1080`, `Listening port`)
	directory := flag.String(`directory`, `/www/`, `Directory to serve`)
	spa := flag.Bool(`spa`, false, `Indicate Single Page Application mode`)
	notFound := flag.Bool(`notFound`, false, `Graceful 404 page at /404.html`)
	flag.StringVar(&domain, `domain`, ``, `Domains names for Content-Security-Policy`)
	flag.Parse()

	if isFileExist(*directory) == nil {
		log.Fatal(`Directory ` + *directory + ` is unreachable.`)
	}

	log.Println(`Starting server on port ` + *port)
	log.Println(`Serving file from ` + *directory)
	log.Println(`Content-Security-Policy: `, contentSecurityPolicy+domain)

	if *spa {
		log.Println(`Working in SPA mode`)
	}

	var notFoundPath *string

	if *notFound {
		if notFoundPath = isFileExist(*directory, NOT_FOUND_FILENAME); notFoundPath == nil {
			log.Println(*directory + NOT_FOUND_FILENAME + ` is unreachable. Flag ignored.`)
			*notFound = false
		} else {
			log.Println(`404 will be ` + *notFoundPath)
		}
	}

	http.Handle(`/`, OwaspHandler{CustomFileHandler{directory, *spa, *notFound, notFoundPath}})

	log.Fatal(http.ListenAndServe(`:`+*port, nil))
}
