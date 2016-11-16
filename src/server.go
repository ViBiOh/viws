package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

const tenDaysOfCaching = `864000`
const contentSecurityPolicy = `default-src 'self' 'unsafe-inline' `
const indexFileName = `index.html`

var directory string
var port string
var domain string
var static string
var notFoundName string
var notFoundPath string

var customNotFound bool
var spa bool

type OwaspMiddleware struct {
	http.ResponseWriter
}

func (m *OwaspMiddleware) WriteHeader(status int) {
	if status < http.StatusBadRequest || customNotFound {
		m.Header().Add(`Content-Security-Policy`, contentSecurityPolicy+domain)
		m.Header().Add(`X-Frame-Options`, `deny`)
		m.Header().Add(`X-Content-Type-Options`, `nosniff`)
		m.Header().Add(`X-XSS-Protection`, `1; mode=block`)
	}
	m.ResponseWriter.WriteHeader(status)
}

type OwaspHandler struct {
	h http.Handler
}

func (handler OwaspHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.h.ServeHTTP(&OwaspMiddleware{ResponseWriter: w}, r)
}

type CustomMiddleware struct {
	http.ResponseWriter
	isNotFound bool
}

func (m *CustomMiddleware) WriteHeader(status int) {
	if status == http.StatusOK || status == http.StatusMovedPermanently {
		m.Header().Add(`Cache-Control`, `max-age=`+tenDaysOfCaching)
	}

	if status == http.StatusNotFound && customNotFound {
		m.isNotFound = true
		m.Header().Add(`Content-type`, `text/html; charset=utf-8`)
	}

	m.ResponseWriter.WriteHeader(status)
}

func (m *CustomMiddleware) Write(p []byte) (int, error) {
	if m.isNotFound {
		notFoundPage, err := ioutil.ReadFile(notFoundPath)
		if err == nil {
			return m.ResponseWriter.Write(notFoundPage)
		}
	}
	return m.ResponseWriter.Write(p)
}

type CustomHandler struct {
	h http.Handler
}

func (handler CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.h.ServeHTTP(&CustomMiddleware{ResponseWriter: w}, r)
}

type SpaMiddleware struct {
	http.ResponseWriter
	isNotFound bool
}

func (m *SpaMiddleware) WriteHeader(status int) {
	if status == http.StatusNotFound {
		m.isNotFound = true
		m.Header().Add(`Content-type`, `text/html; charset=utf-8`)
	}

	m.ResponseWriter.WriteHeader(status)
}

func (m *SpaMiddleware) Write(p []byte) (int, error) {
	if m.isNotFound {
		notFoundPage, err := ioutil.ReadFile(directory + indexFileName)
		if err == nil {
			return m.ResponseWriter.Write(notFoundPage)
		}
	}
	return m.ResponseWriter.Write(p)
}

type SpaHandler struct {
	h http.Handler
}

func (handler SpaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.h.ServeHTTP(&SpaMiddleware{ResponseWriter: w}, r)
}

func checkCustomNotFound() {
	if _, err := os.Stat(notFoundPath); os.IsNotExist(err) {
		log.Println(notFoundPath + ` is not found. Flag ignored.`)
		customNotFound = false
	} else {
		log.Println(`404 page is ` + notFoundPath)
	}
}

func main() {
	flag.BoolVar(&spa, `spa`, false, `Indicate Single Page Application mode`)
	flag.BoolVar(&customNotFound, `notFound`, false, `Graceful 404 page at /404.html`)
	flag.StringVar(&domain, `domain`, ``, `Domains names for Content-Security-Policy`)
	flag.StringVar(&directory, `directory`, `/www/`, `Directory to serve`)
	flag.StringVar(&static, `static`, `/static/`, `Static path served when SPA enabled`)
	flag.StringVar(&notFoundName, `notFoundName`, `404.html`, `Page served when notFound enabled (only for static in SPA)`)
	flag.StringVar(&port, `port`, `1080`, `Listening port`)
	flag.Parse()

	log.Println(`Starting server on port ` + port)
	log.Println(`Content-Security-Policy: `, contentSecurityPolicy+domain)

	pathToServe := `/`
	if spa {
		log.Println(`Working in SPA mode`)
		http.Handle(pathToServe, SpaHandler{OwaspHandler{(http.FileServer(http.Dir(directory)))}})
		pathToServe = static
	}
	http.Handle(pathToServe, CustomHandler{OwaspHandler{(http.FileServer(http.Dir(directory)))}})
	log.Println(`Serving file from ` + path.Join(directory, pathToServe))

	if customNotFound {
		notFoundPath = path.Join(directory, pathToServe, notFoundName)
		checkCustomNotFound()
	}

	log.Fatal(http.ListenAndServe(`:`+port, nil))
}
