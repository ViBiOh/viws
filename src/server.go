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

func (middleware *OwaspMiddleware) WriteHeader(status int) {
	if status < http.StatusBadRequest || customNotFound {
		middleware.Header().Add("Content-Security-Policy", contentSecurityPolicy+domain)
		middleware.Header().Add("X-Frame-Options", "deny")
		middleware.Header().Add("X-Content-Type-Options", "nosniff")
		middleware.Header().Add("X-XSS-Protection", "1; mode=block")
	}
	middleware.ResponseWriter.WriteHeader(status)
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

func (middleware *CustomMiddleware) WriteHeader(status int) {
	if status == http.StatusOK || status == http.StatusMovedPermanently {
		middleware.Header().Add("Cache-Control", "max-age="+tenDaysOfCaching)
	}

	if status == http.StatusNotFound && customNotFound {
		middleware.isNotFound = true
		middleware.Header().Add("Content-type", "text/html; charset=utf-8")
	}

	middleware.ResponseWriter.WriteHeader(status)
}

func (middleware *CustomMiddleware) Write(p []byte) (int, error) {
	if middleware.isNotFound {
		notFoundPage, err := ioutil.ReadFile(notFoundPath)
		if err == nil {
			return middleware.ResponseWriter.Write(notFoundPage)
		}
	}
	return middleware.ResponseWriter.Write(p)
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

func (middleware IndexMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
