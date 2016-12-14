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
65
66
67
68
69
70
71
72
73
74
75
76
77
78
79
80
81
82
83
84
85
86
87
88
89
90
91
92
93
94
95
96
97
98
99
100
101
102
103
104
105
106
107
108
109
110
111
112
113
114
115
116
117
118
119

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
)

const tenDaysOfCaching = `864000`
const contentSecurityPolicy = `default-src 'self' 'unsafe-inline' `
const notFoundFilename = `404.html`
const indexFilename = `index.html`

var domain string

func isFileExist(parts ...string) *string {
	fullPath := path.Join(parts...)
	info, err := os.Stat(fullPath)

	if err != nil {
		return nil
	}

	if info.IsDir() {
		if isFileExist(append(parts, indexFilename)...) == nil {
			return nil
		}
	}

	return &fullPath
}

type owaspMiddleware struct {
	http.ResponseWriter
}

func (m *owaspMiddleware) WriteHeader(status int) {
	if status < http.StatusBadRequest {
		m.Header().Add(`Content-Security-Policy`, contentSecurityPolicy+domain)
		m.Header().Add(`X-Frame-Options`, `deny`)
		m.Header().Add(`X-Content-Type-Options`, `nosniff`)
		m.Header().Add(`X-XSS-Protection`, `1; mode=block`)
	}

	if status == http.StatusOK || status == http.StatusMovedPermanently {
		m.Header().Add(`Cache-Control`, `max-age=`+tenDaysOfCaching)
	}

	m.ResponseWriter.WriteHeader(status)
}

type OwaspHandler struct {
	h http.Handler
}

func (handler OwaspHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.h.ServeHTTP(&owaspMiddleware{ResponseWriter: w}, r)
}

type customFileHandler struct {
	root         *string
	spa          bool
	notFound     bool
	notFoundPath *string
}

func (handler customFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if filePath := isFileExist(*handler.root, r.URL.Path); filePath != nil {
		http.ServeFile(w, r, *filePath)
	} else if handler.notFound {
		http.ServeFile(w, r, *handler.notFoundPath)
	} else if handler.spa {
		http.ServeFile(w, r, *handler.root)
	} else {
		http.Error(w, `404 page not found: `+r.URL.Path, 404)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	port := flag.String(`port`, `1080`, `Listening port`)
	directory := flag.String(`directory`, `/www/`, `Directory to serve`)
	spa := flag.Bool(`spa`, false, `Indicate Single Page Application mode`)
	notFound := flag.Bool(`notFound`, false, `Graceful 404 page at /404.html`)
	flag.StringVar(&domain, `domain`, ``, `Domains names for Content-Security-Policy`)
	flag.Parse()

	if isFileExist(*directory) == nil {
		log.Fatal(`Directory ` + *directory + ` is unreachable.`)
	}

	log.Println(`Starting server on port ` + *port)
	log.Println(`Serving file from ` + *directory)
	log.Println(`Content-Security-Policy: `, contentSecurityPolicy+domain)

	if *spa {
		log.Println(`Working in SPA mode`)
	}

	var notFoundPath *string

	if *notFound {
		if notFoundPath = isFileExist(*directory, notFoundFilename); notFoundPath == nil {
			log.Println(*directory + notFoundFilename + ` is unreachable. Flag ignored.`)
			*notFound = false
		} else {
			log.Println(`404 will be ` + *notFoundPath)
		}
	}

	http.Handle(`/`, OwaspHandler{customFileHandler{directory, *spa, *notFound, notFoundPath}})

	log.Fatal(http.ListenAndServe(`:`+*port, nil))
}

