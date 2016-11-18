package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
)

const tenDaysOfCaching = `864000`
const contentSecurityPolicy = `default-src 'self' 'unsafe-inline' `

var domain string
var spa bool
var notFound bool
var notFoundPath string

func isFileExist(directory string, pathToTest string) (string, bool) {
	fullPath := path.Join(directory, pathToTest)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fullPath, false
	}
	return fullPath, true
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
	root string
}

func (handler CustomFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if filePath, valid := isFileExist(handler.root, r.URL.Path); valid {
		http.ServeFile(w, r, filePath)
	} else if notFound {
		http.ServeFile(w, r, notFoundPath)
	} else if spa {
		http.ServeFile(w, r, handler.root)
	} else {
		http.Error(w, `404 not found: `+r.URL.Path, 500)
	}
}

func main() {
	port := flag.String(`port`, `1080`, `Listening port`)
	directory := flag.String(`directory`, `/www/`, `Directory to serve`)
	flag.BoolVar(&spa, `spa`, false, `Indicate Single Page Application mode`)
	flag.BoolVar(&notFound, `notFound`, false, `Graceful 404 page at /404.html`)
	flag.StringVar(&domain, `domain`, ``, `Domains names for Content-Security-Policy`)
	flag.Parse()

	log.Println(`Starting server on port ` + *port)
	log.Println(`Serving file from ` + *directory)
	log.Println(`Content-Security-Policy: `, contentSecurityPolicy+domain)

	if spa {
		log.Println(`Working in SPA mode`)
	}

	if notFound {
		if notFoundPath, valid := isFileExist(*directory, `404.html`); !valid {
			log.Println(notFoundPath + ` is not found. Flag ignored.`)
			notFound = false
		}
	}

	http.Handle(`/`, OwaspHandler{CustomFileHandler{*directory}})

	log.Fatal(http.ListenAndServe(`:`+*port, nil))
}
