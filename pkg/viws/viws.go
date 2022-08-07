package viws

import (
	"bytes"
	"errors"
	"flag"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

const (
	indexFilename      = "index.html"
	notFoundFilename   = "404.html"
	cacheControlHeader = "Cache-Control"
	noCacheValue       = "no-cache"
)

var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 1024))
	},
}

// App of package
type App struct {
	headers   http.Header
	directory string
	spa       bool
}

// Config of package
type Config struct {
	directory *string
	headers   *string
	spa       *bool
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		directory: flags.String(fs, prefix, "viws", "Directory", "Directory to serve", "/www/", overrides),
		headers:   flags.String(fs, prefix, "viws", "Headers", "Custom headers, tilde separated (e.g. content-language:fr~X-UA-Compatible:test)", "", overrides),
		spa:       flags.Bool(fs, prefix, "viws", "Spa", "Indicate Single Page Application mode", false, overrides),
	}
}

// New creates new App from Config
func New(config Config) App {
	a := App{
		spa:       *config.spa,
		directory: *config.directory,
		headers:   http.Header{},
	}

	logger.WithField("dir", a.directory).Info("Serving file")

	if a.spa {
		logger.WithField("dir", a.directory).Info("Single Page Application mode enabled")
	}

	if len(*config.headers) != 0 {
		for _, header := range strings.Split(*config.headers, "~") {
			if parts := strings.SplitN(header, ":", 2); len(parts) != 2 || strings.Contains(parts[0], " ") {
				logger.WithField("dir", a.directory).WithField("header", header).Warn("header has wrong format")
			} else {
				a.headers.Add(parts[0], parts[1])
			}
		}
	}

	return a
}

// Handler serve file given configuration
func (a App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if strings.Contains(r.URL.Path, "..") {
			httperror.BadRequest(w, errors.New("path with dots are not allowed"))
		}

		if filename, modTime, err := getFileToServe(a.directory, r.URL.Path); err == nil {
			a.serveFile(w, r, filename, modTime)
			return
		}

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if a.spa {
			if filename, modTime, err := getFileToServe(a.directory, indexFilename); err == nil {
				w.Header().Add(cacheControlHeader, noCacheValue)
				a.serveFile(w, r, filename, modTime)
				return
			}
		}

		a.serveNotFound(w)
	})
}

func (a App) serveFile(w http.ResponseWriter, r *http.Request, filepath string, modTime time.Time) {
	a.addCustomHeaders(w)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	file, err := os.OpenFile(filepath, os.O_RDONLY, 0o600)
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			logger.Error("close file: %s", err)
		}
	}()

	setCacheHeader(w, r)
	http.ServeContent(w, r, filepath, modTime, file)
}

func (a App) addCustomHeaders(w http.ResponseWriter) {
	for key, value := range a.headers {
		w.Header()[key] = value
	}
}

func setCacheHeader(w http.ResponseWriter, r *http.Request) {
	if len(w.Header().Get(cacheControlHeader)) == 0 {
		if r.URL.Path == "/" {
			w.Header().Add(cacheControlHeader, noCacheValue)
		} else {
			w.Header().Add(cacheControlHeader, "public, max-age=864000")
		}
	}
}
