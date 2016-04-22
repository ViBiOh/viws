package main

import "net/http"
import "log"
import "compress/gzip"
import "io"
import "strings"
import "strconv"

const port = "1080"
const directory = "/www/"
const tenDaysOfCaching = "864000"

type OwaspHeaderServer struct {
	http.ResponseWriter
}

func (w OwaspHeaderServer) WriteHeader(code int) {
	if code == 200 {
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

type GzipServer struct {
	io.Writer
	http.ResponseWriter
}

func (w GzipServer) Write(b []byte) (int, error) {
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	return w.Writer.Write(b)
}

func (w GzipServer) WriteHeader(code int) {
	if code == 200 {
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
	}

	w.ResponseWriter.WriteHeader(code)
}

func gzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		gz := gzip.NewWriter(w)
		defer gz.Close()

		h.ServeHTTP(GzipServer{ResponseWriter: w, Writer: gz}, r)
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
		log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.Path)

		if len(r.URL.Path) > 1 && strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		h.ServeHTTP(CustomFileServer{ResponseWriter: w}, r)
	})
}

func main() {
	http.Handle("/", customMiddleware(owaspMiddleware(gzipMiddleware(http.FileServer(http.Dir(directory))))))

	log.Println("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
