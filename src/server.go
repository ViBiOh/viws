package main

import "net/http"
import "log"

const port = "1080"
const directory = "/www/"

type CustomFileServer struct {
	http.ResponseWriter
}

func (w CustomFileServer) WriteHeader(code int) {
	if code == 200 {
		w.Header().Add("Cache-Control", "max-age=3153600")
	}
	w.ResponseWriter.WriteHeader(code)
}

func customFileServer(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(CustomFileServer{w}, r)
	})
}

func main() {
	http.Handle("/", customFileServer(http.FileServer(http.Dir(directory))))

	log.Print("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
