package main

import "net/http"
import "log"
import "os"
import "path"
import "flag"
import "io/ioutil"

const tenDaysOfCaching = "864000"
const contentSecurityPolicy = "default-src 'self' 'unsafe-inline' "

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

func (w *OwaspMiddleware) WriteHeader(status int) {
	if status < http.StatusBadRequest || customNotFound {
		w.Header().Add("Content-Security-Policy", contentSecurityPolicy+domain)
		w.Header().Add("X-Frame-Options", "deny")
		w.Header().Add("X-Content-Type-Options", "nosniff")
		w.Header().Add("X-XSS-Protection", "1; mode=block")
	}
	w.ResponseWriter.WriteHeader(status)
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

func (w *CustomMiddleware) WriteHeader(status int) {
	if status == http.StatusOK || status == http.StatusMovedPermanently {
		w.Header().Add("Cache-Control", "max-age="+tenDaysOfCaching)
	}

	if status == http.StatusNotFound && customNotFound {
		w.isNotFound = true
		w.Header().Add("Content-type", "text/html; charset=utf-8")
	}

	w.ResponseWriter.WriteHeader(status)
}

func (w *CustomMiddleware) Write(p []byte) (int, error) {
	if w.isNotFound {
		notFoundPage, err := ioutil.ReadFile(notFoundPath)
		if err == nil {
			return w.ResponseWriter.Write(notFoundPage)
		}
	}
	return w.ResponseWriter.Write(p)
}

type CustomHandler struct {
	h http.Handler
}

func (handler CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.h.ServeHTTP(&CustomMiddleware{ResponseWriter: w}, r)
}

type IndexMiddleware struct {
	http.ResponseWriter
	http.Handler
}

func (m IndexMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, directory)
}

func checkCustomNotFound() {
	if _, err := os.Stat(notFoundPath); os.IsNotExist(err) {
		log.Println(notFoundPath + " is not found. Flag ignored.")
		customNotFound = false
	} else {
		log.Println("404 page is " + notFoundPath)
	}
}

func main() {
	flag.BoolVar(&spa, "spa", false, "Indicate Single Page Application mode")
	flag.BoolVar(&customNotFound, "notFound", false, "Graceful 404 page at /404.html")
	flag.StringVar(&domain, "domain", "", "Domains names for Content-Security-Policy")
	flag.StringVar(&directory, "directory", "/www/", "Directory to serve")
	flag.StringVar(&static, "static", "/static/", "Static path served when SPA enabled")
	flag.StringVar(&notFoundName, "notFoundName", "404.html", "Page served when notFound enabled")
	flag.StringVar(&port, "port", "1080", "Listening port")
	flag.Parse()

	log.Println("Starting server on port " + port)
	log.Println("Content-Security-Policy: ", contentSecurityPolicy+domain)

	pathToServe := "/"
	if spa {
		log.Println("Working in SPA mode")
		http.Handle(pathToServe, CustomHandler{OwaspHandler{(IndexMiddleware{})}})
		pathToServe = static
	}
	http.Handle(pathToServe, CustomHandler{OwaspHandler{(http.FileServer(http.Dir(directory)))}})
	log.Println("Serving file from " + path.Join(directory, pathToServe))

	if customNotFound {
		notFoundPath = path.Join(directory, pathToServe, notFoundName)
		checkCustomNotFound()
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
