package viws

import (
	"bytes"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/hash"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
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

type App struct {
	headers   http.Header
	directory string
	spa       bool
}

type Config struct {
	Directory string
	Headers   []string
	Spa       bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Directory", "Directory to serve").Prefix(prefix).DocPrefix("viws").StringVar(fs, &config.Directory, "/www/", overrides)
	flags.New("Header", "Custom header e.g. content-language:fr").Prefix(prefix).DocPrefix("viws").StringSliceVar(fs, &config.Headers, nil, overrides)
	flags.New("Spa", "Indicate Single Page Application mode").Prefix(prefix).DocPrefix("viws").BoolVar(fs, &config.Spa, false, overrides)

	return &config
}

func New(config *Config) App {
	a := App{
		spa:       config.Spa,
		directory: config.Directory,
		headers:   http.Header{},
	}

	logger := slog.With("dir", a.directory)

	logger.Info("Serving file")

	if a.spa {
		logger.Info("Single Page Application mode enabled")
	}

	if len(config.Headers) != 0 {
		for _, header := range config.Headers {
			if parts := strings.SplitN(header, ":", 2); len(parts) != 2 || strings.Contains(parts[0], " ") {
				logger.Warn("header has wrong format", "header", header)
			} else {
				a.headers.Add(parts[0], parts[1])
			}
		}
	}

	return a
}

func (a App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") {
			httperror.BadRequest(r.Context(), w, fmt.Errorf("path with dots are not allowed: `%s`", r.URL.Path))
			return
		}

		if filename, info, err := getFileToServe(a.directory, r.URL.Path); err == nil {
			a.serveFile(w, r, filename, hash.Hash(info), info.ModTime())
			return
		}

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if a.spa {
			if filename, info, err := getFileToServe(a.directory, indexFilename); err == nil {
				w.Header().Add(cacheControlHeader, noCacheValue)
				a.serveFile(w, r, filename, hash.Hash(info), info.ModTime())
				return
			}
		}

		a.serveNotFound(r.Context(), w)
	})
}

func (a App) serveFile(w http.ResponseWriter, r *http.Request, filepath, hash string, modTime time.Time) {
	a.addCustomHeaders(w)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	etag, ok := etagMatch(w, r, hash)
	if ok {
		return
	}

	file, err := os.OpenFile(filepath, os.O_RDONLY, 0o600)
	if err != nil {
		httperror.InternalServerError(r.Context(), w, err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			slog.LogAttrs(r.Context(), slog.LevelError, "close file", slog.Any("error", err))
		}
	}()

	setCacheHeader(w, r)
	w.Header().Add("Etag", etag)
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

func etagMatch(w http.ResponseWriter, r *http.Request, hash string) (etag string, match bool) {
	etag = fmt.Sprintf(`W/"%s"`, hash)

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		match = true
	}

	return
}
