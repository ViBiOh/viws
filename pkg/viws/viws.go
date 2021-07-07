package viws

import (
	"bytes"
	"flag"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/query"
)

const (
	notFoundFilename   = "404.html"
	cacheControlHeader = "Cache-Control"
	noCacheValue       = "no-cache"
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 32*1024))
		},
	}
)

// Config of package
type Config struct {
	directory *string
	headers   *string
	spa       *bool
}

// App of package
type App interface {
	Handler() http.Handler
}

type app struct {
	headers   map[string]string
	directory string
	spa       bool
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		directory: flags.New(prefix, "viws").Name("Directory").Default(flags.Default("Directory", "/www/", overrides)).Label("Directory to serve").ToString(fs),
		headers:   flags.New(prefix, "viws").Name("Headers").Default(flags.Default("Headers", "", overrides)).Label("Custom headers, tilde separated (e.g. content-language:fr~X-UA-Compatible:test)").ToString(fs),
		spa:       flags.New(prefix, "viws").Name("Spa").Default(flags.Default("Spa", false, overrides)).Label("Indicate Single Page Application mode").ToBool(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	a := app{
		spa:       *config.spa,
		directory: strings.TrimSpace(*config.directory),
	}

	logger.WithField("dir", a.directory).Info("Serving file")

	if a.spa {
		logger.Info("Single Page Application mode enabled")
	}

	rawHeaders := strings.TrimSpace(*config.headers)
	if len(rawHeaders) != 0 {
		a.headers = make(map[string]string)

		for _, header := range strings.Split(rawHeaders, "~") {
			if parts := strings.SplitN(header, ":", 2); len(parts) != 2 || strings.Contains(parts[0], " ") {
				logger.WithField("header", header).Warn("header has wrong format")
			} else {
				a.headers[parts[0]] = parts[1]
			}
		}
	}

	return a
}

func (a app) addCustomHeaders(w http.ResponseWriter) {
	for key, value := range a.headers {
		w.Header().Add(key, value)
	}
}

func setCacheHeader(w http.ResponseWriter, r *http.Request) {
	if len(w.Header().Get(cacheControlHeader)) == 0 {
		if query.IsRoot(r) {
			w.Header().Set(cacheControlHeader, noCacheValue)
		} else {
			w.Header().Set(cacheControlHeader, "public, max-age=864000")
		}
	}
}

func (a app) serveFile(w http.ResponseWriter, r *http.Request, filepath string) {
	a.addCustomHeaders(w)

	if r.Method == http.MethodGet {
		setCacheHeader(w, r)
		http.ServeFile(w, r, filepath)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (a app) serveNotFound(w http.ResponseWriter) {
	notFoundPath, err := getFileToServe(a.directory, notFoundFilename)
	if os.IsNotExist(err) {
		httperror.NotFound(w)
		return
	}

	file, err := os.Open(notFoundPath)
	if file != nil {
		defer func() {
			if err := file.Close(); err != nil {
				logger.Error("unable to close file: %s", err)
			}
		}()
	}
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}

	contentType := mime.TypeByExtension(notFoundPath)
	if len(contentType) == 0 {
		contentType = "text/html; charset=utf-8"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set(cacheControlHeader, noCacheValue)
	w.WriteHeader(http.StatusNotFound)

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	if _, err = io.CopyBuffer(w, file, buffer.Bytes()); err != nil {
		logger.Error("unable to copy content to writer: %s", err)
	}
}

// Handler serve file given configuration
func (a app) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if filename, err := getFileToServe(a.directory, r.URL.Path); err == nil {
			a.serveFile(w, r, filename)
			return
		}

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if a.spa {
			w.Header().Set(cacheControlHeader, noCacheValue)
			a.serveFile(w, r, path.Join(a.directory, "index.html"))
			return
		}

		a.serveNotFound(w)
	})
}
