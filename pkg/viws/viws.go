package viws

import (
	"flag"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/httputils/v2/pkg/errors"
	"github.com/ViBiOh/httputils/v2/pkg/httperror"
	"github.com/ViBiOh/httputils/v2/pkg/logger"
	"github.com/ViBiOh/httputils/v2/pkg/tools"
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
	spa          bool
	directory    string
	pushPaths    []string
	headers      map[string]string
	notFoundPath string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		directory: tools.NewFlag(prefix, "viws").Name("Directory").Default("/www/").Label("Directory to serve").ToString(fs),
		headers:   tools.NewFlag(prefix, "viws").Name("Headers").Default("").Label("Custom headers, tilde separated (e.g. content-language:fr~X-UA-Compatible:test)").ToString(fs),
		spa:       tools.NewFlag(prefix, "viws").Name("Spa").Default(false).Label("Indicate Single Page Application mode").ToBool(fs),
		push:      tools.NewFlag(prefix, "viws").Name("Push").Default("").Label("Paths for HTTP/2 Server Push on index, comma separated").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) (App, error) {
	directory := strings.TrimSpace(*config.directory)
	if _, err := getFileToServe(directory); err != nil {
		return nil, errors.Wrap(err, "directory %s is unreachable or does not contains index", directory)
	}
	logger.Info("Serving file from %s", directory)

	spa := *config.spa
	var notFoundPath string

	if !spa {
		if path, err := getFileToServe(directory, notFoundFilename); err != nil {
			logger.Warn("no %s file on directory, 404 will be plain text", notFoundFilename)
		} else {
			notFoundPath = path
		}
	}

	var pushPaths []string
	push := strings.TrimSpace(*config.push)

	if push != "" {
		pushPaths = strings.Split(push, ",")
		logger.Info("HTTP/2 Push of %s", pushPaths)
	}

	var headers map[string]string
	rawHeaders := strings.TrimSpace(*config.headers)

	if rawHeaders != "" {
		headers = make(map[string]string)

		for _, header := range strings.Split(rawHeaders, "~") {
			if parts := strings.SplitN(header, ":", 2); len(parts) != 2 {
				logger.Warn("header has wrong format: %s", header)
			} else {
				headers[parts[0]] = parts[1]
			}
		}
	}

	if *config.spa {
		logger.Info("Single Page Application mode enabled")
	}

	return app{
		spa:          spa,
		directory:    directory,
		pushPaths:    pushPaths,
		notFoundPath: notFoundPath,
		headers:      headers,
	}, nil
}

func (a app) addCustomHeaders(w http.ResponseWriter) {
	for key, value := range a.headers {
		w.Header().Add(key, value)
	}
}

func (a app) handlePush(w http.ResponseWriter, r *http.Request) {
	if pusher, ok := w.(http.Pusher); ok {
		for _, path := range a.pushPaths {
			if err := pusher.Push(path, nil); err != nil {
				logger.Error("failed to push %s: %#v", path, err)
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

func (a app) serveLocalFile(w http.ResponseWriter, filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		httperror.InternalServerError(w, errors.WithStack(err))
		return
	}

	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath))
	if _, err = io.Copy(w, file); err != nil {
		logger.Error("unable to copy content to writer: %#v", errors.WithStack(err))
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

		if hasPush && r.Method == http.MethodGet && tools.IsRoot(r) {
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

		if a.notFoundPath != "" {
			a.serveLocalFile(w, a.notFoundPath)
			return
		}

		httperror.NotFound(w)
	})
}
