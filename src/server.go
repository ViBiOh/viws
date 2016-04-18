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
	io.Writer
	http.ResponseWriter
}

func (w CustomFileServer) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w CustomFileServer) WriteHeader(code int) {
	if code == 200 {
		w.Header().Add("Content-Security-Policy", "default-src 'self' https://apis.google.com")
		w.Header().Add("X-Frame-Options", "deny")
		w.Header().Add("X-Content-Type-Options", "nosniff")
		w.Header().Add("X-XSS-Protection", "1; mode=block")
		w.Header().Add("Cache-Control", "max-age="+tenDaysOfCaching)
		w.Header().Add("Vary", "Accept-Encoding")
	}
	w.ResponseWriter.WriteHeader(code)
}

func customFileServer(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(CustomFileServer{ResponseWriter: w}, r)
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		h.ServeHTTP(CustomFileServer{ResponseWriter: w, Writer: gz}, r)
	})
}

func main() {
	http.Handle("/", customFileServer(http.FileServer(http.Dir(directory))))

	log.Print("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
