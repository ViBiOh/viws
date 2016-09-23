package main

import "net/http"
import "log"
import "os"
import "flag"
import "io/ioutil"

const port = "1080"
const directory = "/www/"
const staticPath = "/static/"
const notFoundName = "404.html"
const tenDaysOfCaching = "864000"
const contentSecurityPolicy = "default-src 'self' 'unsafe-inline' "

var customNotFound bool
var notFoundPath string
var domain string
var spa bool

type OwaspHeaderServer struct {
	http.ResponseWriter
}

func (w OwaspHeaderServer) WriteHeader(code int) {
	if code < 400 || customNotFound {
		w.Header().Add("Content-Security-Policy", contentSecurityPolicy+domain)
		w.Header().Add("X-Frame-Options", "deny")
		w.Header().Add("X-Content-Type-Options", "nosniff")
		w.Header().Add("X-XSS-Protection", "1; mode=block")
	}
	w.ResponseWriter.WriteHeader(code)
}

func owaspMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(OwaspHeaderServer{ResponseWriter: w}, r)
	})
}

type CustomFileServer struct {
	http.ResponseWriter
	isNotFound bool
}

func (w *CustomFileServer) WriteHeader(code int) {
	if code == 200 || code == 301 {
		w.Header().Add("Cache-Control", "max-age="+tenDaysOfCaching)
	}

	if code == 404 {
		w.isNotFound = true
		if customNotFound {
			w.Header().Add("Content-type", "text/html; charset=utf-8")
		}
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *CustomFileServer) Write(p []byte) (int, error) {
	if w.isNotFound && customNotFound {
		notFoundPage, err := ioutil.ReadFile(notFoundPath)
		if err == nil {
			return w.ResponseWriter.Write(notFoundPage)
		}
	}
	return w.ResponseWriter.Write(p)
}

type CustomFileServerHandler struct {
	h http.Handler
}

func (handler CustomFileServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.h.ServeHTTP(&CustomFileServer{ResponseWriter: w}, r)
}

type IndexMiddleware struct {
	http.ResponseWriter
	http.Handler
}

func (m IndexMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, directory)
}

func main() {
	flag.BoolVar(&spa, "spa", false, "Indicate Single Page Application mode")
	flag.BoolVar(&customNotFound, "notFound", false, "Graceful 404 page at /404.html")
	flag.StringVar(&domain, "domain", "", "Domains names for Content-Security-Policy")
	flag.Parse()

	pathToServe := "/"
	if spa {
		log.Println("Working in SPA mode")
		http.Handle(pathToServe, CustomFileServerHandler{(owaspMiddleware(IndexMiddleware{}))})
		pathToServe = staticPath
	}
	notFoundPath = directory + pathToServe + notFoundName
	http.Handle(pathToServe, CustomFileServerHandler{(owaspMiddleware(http.FileServer(http.Dir(directory))))})

	if customNotFound {
		if _, err := os.Stat(notFoundPath); os.IsNotExist(err) {
			log.Println(notFoundPath + " is not found. Flag ignored")
			customNotFound = false
		} else {
			log.Println("404 page is " + notFoundPath)
		}
	}

	log.Println("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
