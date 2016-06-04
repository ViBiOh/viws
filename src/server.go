package main

import "net/http"
import "log"
import "flag"

const port = "1080"
const directory = "/Users/vibioh/code/angular-bp/dist/"
const tenDaysOfCaching = "864000"

type OwaspHeaderServer struct {
	http.ResponseWriter
}

func (w OwaspHeaderServer) WriteHeader(code int) {
	if code < 400 {
		w.Header().Add("Content-Security-Policy", "default-src 'self' 'unsafe-inline' http://*.vibioh.fr https://*.vibioh.fr https://apis.google.com https://fonts.googleapis.com https://fonts.gstatic.com")
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
}

func (w CustomFileServer) WriteHeader(code int) {
	if code == 200 {
		w.Header().Add("Cache-Control", "max-age="+tenDaysOfCaching)
	}

	w.ResponseWriter.WriteHeader(code)
}

func customMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(CustomFileServer{ResponseWriter: w}, r)
	})
}

type IndexMiddleware struct {
	http.ResponseWriter
	http.Handler
}

func (m IndexMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, directory)
}

func main() {
	spa := flag.Bool("spa", false, "Indicate Single Page Application mode")
	flag.Parse()

	pathToServe := "/"
	if (*spa) {
		log.Println("Working in SPA mode")
		pathToServe = "/static/"
		http.Handle("/", customMiddleware(owaspMiddleware(IndexMiddleware{})))
	}
	http.Handle(pathToServe, customMiddleware(owaspMiddleware(http.FileServer(http.Dir(directory)))))

	log.Println("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
