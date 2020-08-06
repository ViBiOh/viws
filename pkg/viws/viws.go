package viws

import (
	"flag"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/query"
)

const (
	notFoundFilename = "404.html"
)

// Config of package
type Config struct {
	directory *string
	headers   *string
	spa       *bool
	push      *string
}

// App of package
type App interface {
	Handler() http.Handler
}

type app struct {
	spa       bool
	directory string
	pushPaths []string
	headers   map[string]string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		directory: flags.New(prefix, "viws").Name("Directory").Default("/www/").Label("Directory to serve").ToString(fs),
		headers:   flags.New(prefix, "viws").Name("Headers").Default("").Label("Custom headers, tilde separated (e.g. content-language:fr~X-UA-Compatible:test)").ToString(fs),
		spa:       flags.New(prefix, "viws").Name("Spa").Default(false).Label("Indicate Single Page Application mode").ToBool(fs),
		push:      flags.New(prefix, "viws").Name("Push").Default("").Label("Paths for HTTP/2 Server Push on index, comma separated").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	a := app{
		spa:       *config.spa,
		directory: strings.TrimSpace(*config.directory),
	}

	logger.Info("Serving file from %s", a.directory)

	if a.spa {
		logger.Info("Single Page Application mode enabled")
	}

	push := strings.TrimSpace(*config.push)
	if len(push) != 0 {
		a.pushPaths = strings.Split(push, ",")
		logger.Info("HTTP/2 Push of %s", strings.Join(a.pushPaths, ", "))
	}

	rawHeaders := strings.TrimSpace(*config.headers)
	if len(rawHeaders) != 0 {
		a.headers = make(map[string]string)

		for _, header := range strings.Split(rawHeaders, "~") {
			if parts := strings.SplitN(header, ":", 2); len(parts) != 2 || strings.Contains(parts[0], " ") {
				logger.Warn("header has wrong format: `%s`", header)
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

func (a app) handlePush(w http.ResponseWriter, _ *http.Request) {
	if pusher, ok := w.(http.Pusher); ok {
		for _, pushPath := range a.pushPaths {
			if err := pusher.Push(pushPath, nil); err != nil {
				logger.Error("failed to push %s: %s", pushPath, err)
			}
		}
	}
}

func (a app) serveFile(w http.ResponseWriter, r *http.Request, filepath string) {
	a.addCustomHeaders(w)

	if r.Method == http.MethodGet {
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
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", mime.TypeByExtension(notFoundPath))
	if _, err = io.Copy(w, file); err != nil {
		logger.Error("unable to copy content to writer: %s", err)
	}
}

// Handler serve file given configuration
func (a app) Handler() http.Handler {
	hasPush := len(a.pushPaths) != 0

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if hasPush && r.Method == http.MethodGet && query.IsRoot(r) {
			a.handlePush(w, r)
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
			w.Header().Set("Cache-Control", "no-cache")
			a.serveFile(w, r, path.Join(a.directory, "index.html"))
			return
		}

		a.serveNotFound(w)
	})
}
