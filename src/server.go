The Go PlaygroundRun  Format   Imports  Share About
1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
18
19
20
21
22
23
24
25
26
27
28
29
30
31
32
33
34
35
36
37
38
39
40
41
42
43
44
45
46
47
48
49
50
51
52
53
54
55
56
57
58
59
60
61
62
63
64

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
		w.Header().Add("Content-Security-Policy", "default-src 'self' 'unsafe-inline' http://*.vibioh.fr https://*.vibioh.fr https://apis.google.com https://fonts.googleapis.com https://fonts.gstatic.com")
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
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(CustomFileServer{ResponseWriter: w}, r)
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		h.ServeHTTP(CustomFileServer{ResponseWriter: w, Writer: gz}, r)
	})
}

func redirectIndex(w http.ResponseWriter, r *http.Request) {
}

func main() {
	http.Handle("/", customFileServer(http.FileServer(http.Dir(directory))))
	http.NotFound = redirectIndex

	log.Print("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
