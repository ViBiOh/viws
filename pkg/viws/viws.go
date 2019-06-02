package viws

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/tools"
)

const (
	notFoundFilename = "404.html"
)

// Config of package
type Config struct {
	directory *string
	headers   *string
	notFound  *bool
	spa       *bool
	push      *string
}

// App of package
type App struct {
	spa          bool
	directory    string
	pushPaths    []string
	headers      map[string]string
	notFoundPath string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	docPrefix := prefix
	if prefix == "" {
		docPrefix = "viws"
	}

	return Config{
		directory: fs.String(tools.ToCamel(fmt.Sprintf("%sDirectory", prefix)), "/www/", fmt.Sprintf("[%s] Directory to serve", docPrefix)),
		headers:   fs.String(tools.ToCamel(fmt.Sprintf("%sHeaders", prefix)), "", fmt.Sprintf("[%s] Custom headers, tilde separated (e.g. content-language:fr~X-UA-Compatible:test)", docPrefix)),
		notFound:  fs.Bool(tools.ToCamel(fmt.Sprintf("%sNotFound", prefix)), false, fmt.Sprintf("[%s] Graceful 404 page at /%s", docPrefix, notFoundFilename)),
		spa:       fs.Bool(tools.ToCamel(fmt.Sprintf("%sSpa", prefix)), false, fmt.Sprintf("[%s] Indicate Single Page Application mode", docPrefix)),
		push:      fs.String(tools.ToCamel(fmt.Sprintf("%sPush", prefix)), "", fmt.Sprintf("[%s] Paths for HTTP/2 Server Push on index, comma separated", docPrefix)),
	}
}

// New creates new App from Config
func New(config Config) (*App, error) {
	if *config.notFound && *config.spa {
		return nil, errors.New("incompatible options provided: -notFound and -spa")
	}

	directory := strings.TrimSpace(*config.directory)
	if _, err := getFileToServe(directory); err != nil {
		return nil, errors.Wrap(err, "directory %s is unreachable or does not contains index", directory)
	}
	logger.Info("Serving file from %s", directory)

	var notFoundPath string
	var err error

	if *config.notFound {
		if notFoundPath, err = getFileToServe(directory, notFoundFilename); err != nil {
			return nil, errors.Wrap(err, "not found page %s is unreachable", notFoundPath)
		}

		logger.Info("404 will be %s", notFoundPath)
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

	return &App{
		spa:          *config.spa,
		directory:    directory,
		pushPaths:    pushPaths,
		notFoundPath: notFoundPath,
		headers:      headers,
	}, nil
}

func (a App) addCustomHeaders(w http.ResponseWriter) {
	for key, value := range a.headers {
		w.Header().Set(key, value)
	}
}

func (a App) handlePush(w http.ResponseWriter, r *http.Request) {
	if pusher, ok := w.(http.Pusher); ok {
		for _, path := range a.pushPaths {
			if err := pusher.Push(path, nil); err != nil {
				logger.Error("failed to push %s: %#v", path, err)
			}
		}
	}
}

func (a App) serve(w http.ResponseWriter, r *http.Request, path string) {
	a.addCustomHeaders(w)

	if r.Method == http.MethodGet {
		http.ServeFile(w, r, path)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler serve file given configuration
func (a App) Handler() http.Handler {
	hasPush := len(a.pushPaths) != 0
	hasNotFound := a.notFoundPath != ""

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if hasPush && r.Method == http.MethodGet && tools.IsRoot(r) {
			a.handlePush(w, r)
		}

		if filename, err := getFileToServe(a.directory, r.URL.Path); err == nil {
			a.serve(w, r, filename)
			return
		}

		if hasNotFound {
			w.WriteHeader(http.StatusNotFound)

			a.serve(w, r, a.notFoundPath)
			return
		}

		if a.spa {
			w.Header().Set("Cache-Control", "no-cache")

			a.serve(w, r, a.directory)
			return
		}

		httperror.NotFound(w)
	})
}
