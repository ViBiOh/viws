package main

import "net/http"
import "log"
import "compress/gzip"
import "io"
import "strings"

const port = "1080"
const directory = "/www/"
const tenDaysOfCaching = "864000"

type CustomFileServer struct {
	http.ResponseWriter
}

func (w CustomFileServer) WriteHeader(code int) {
	if code == 200 {
		w.Header().Add("Content-Security-Policy", "default-src 'self' 'unsafe-inline' http://*.vibioh.fr https://*.vibioh.fr https://apis.google.com https://fonts.googleapis.com https://fonts.gstatic.com")
		w.Header().Add("X-Frame-Options", "deny")
		w.Header().Add("X-Content-Type-Options", "nosniff")
		w.Header().Add("X-XSS-Protection", "1; mode=block")
		w.Header().Add("Cache-Control", "max-age="+tenDaysOfCaching)
	}
	
	w.ResponseWriter.WriteHeader(code)
}

func customFileServer(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.Path)
		
		if len(r.URL.Path) > 1 && strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		h.ServeHTTP(CustomFileServer{ResponseWriter: w}, r)
	})
}

type GzipFilerServer struct {
	io.Writer
	http.ResponseWriter
}

func (w GzipFilerServer) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w GzipFilerServer) WriteHeader(code int) {
	if code == 200 {
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
	}
	
	w.ResponseWriter.WriteHeader(code)
}

func gzipFileHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
		}

		gz := gzip.NewWriter(w)
		defer gz.Close()

		h.ServeHTTP(GzipFilerServer{ResponseWriter: w, Writer: gz}, r)
	})
}

func main() {
	http.Handle("/", customFileServer(gzipFileHandler(http.FileServer(http.Dir(directory)))))

	log.Println("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
