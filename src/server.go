package main

import "net/http"
import "log"
import "compress/gzip"
import "io"
import "strings"

const port = "1080"
const directory = "/www/"

type CustomFileServer struct {
	io.Writer
	http.ResponseWriter
}

func (w CustomFileServer) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w CustomFileServer) WriteHeader(code int) {
	if code == 200 {
		w.Header().Add("Cache-Control", "max-age=3153600")
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
