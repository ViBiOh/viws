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
	"github.com/ViBiOh/viws/pkg/utils"
)

const (
	notFoundFilename = `404.html`
)

// App stores informations and secret of API
type App struct {
	spa          bool
	directory    string
	pushPaths    []string
	headers      map[string]string
	notFoundPath *string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}) (*App, error) {
	spa := *(config[`spa`].(*bool))
	notFound := *(config[`notFound`].(*bool))
	directory := strings.TrimSpace(*(config[`directory`].(*string)))
	push := strings.TrimSpace(*(config[`push`].(*string)))
	rawHeaders := strings.TrimSpace(*(config[`headers`].(*string)))

	if utils.IsFileExist(directory) == nil {
		return nil, errors.New(`directory %s is unreachable or does not contains index`, directory)
	}
	logger.Info(`Serving file from %s`, directory)

	var notFoundPath *string
	if notFound {
		if spa {
			return nil, errors.New(`incompatible options provided: -notFound and -spa`)
		}

		if notFoundPath = utils.IsFileExist(directory, notFoundFilename); notFoundPath == nil {
			return nil, errors.New(`not found page %s%s is unreachable`, directory, notFoundFilename)
		}

		logger.Info(`404 will be %s`, *notFoundPath)
	}

	var pushPaths []string
	if push != `` {
		pushPaths = strings.Split(push, `,`)
		logger.Info(`HTTP/2 Push of %s`, pushPaths)
	}

	headers := make(map[string]string)
	if rawHeaders != `` {
		for _, header := range strings.Split(rawHeaders, `~`) {
			if parts := strings.SplitN(header, `:`, 2); len(parts) != 2 {
				logger.Warn(`header has wrong format: %s`, header)
			} else {
				headers[parts[0]] = parts[1]
			}
		}
	}

	if spa {
		logger.Info(`Working in SPA mode`)
	}

	return &App{
		spa:          spa,
		directory:    directory,
		pushPaths:    pushPaths,
		notFoundPath: notFoundPath,
		headers:      headers,
	}, nil
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`directory`: flag.String(tools.ToCamel(fmt.Sprintf(`%sDirectory`, prefix)), `/www/`, `[viws] Directory to serve`),
		`headers`:   flag.String(tools.ToCamel(fmt.Sprintf(`%sHeaders`, prefix)), ``, `[viws] Custom headers, tilde separated (e.g. content-language:fr~X-UA-Compatible:test)`),
		`notFound`:  flag.Bool(tools.ToCamel(fmt.Sprintf(`%sNotFound`, prefix)), false, `[viws] Graceful 404 page at /404.html`),
		`spa`:       flag.Bool(tools.ToCamel(fmt.Sprintf(`%sSpa`, prefix)), false, `[viws] Indicate Single Page Application mode`),
		`push`:      flag.String(tools.ToCamel(fmt.Sprintf(`%sPush`, prefix)), ``, `[viws] Paths for HTTP/2 Server Push on index, comma separated`),
	}
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
				logger.Error(`failed to push %s: %+v`, path, err)
			}
		}
	}
}

// Handler serve file given configuration
func (a App) Handler() http.Handler {
	hasPush := len(a.pushPaths) != 0

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			a.addCustomHeaders(w)

			if utils.IsFileExist(a.directory, r.URL.Path) != nil {
				w.WriteHeader(http.StatusNoContent)
			} else {
				httperror.NotFound(w)
			}

			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if hasPush && r.URL.Path == `/` {
			a.handlePush(w, r)
		}

		if filename := utils.IsFileExist(a.directory, r.URL.Path); filename != nil {
			a.addCustomHeaders(w)
			http.ServeFile(w, r, *filename)
			return
		}

		if a.notFoundPath != nil {
			w.WriteHeader(http.StatusNotFound)
			a.addCustomHeaders(w)
			http.ServeFile(w, r, *a.notFoundPath)
			return
		}

		if a.spa {
			w.Header().Set(`Cache-Control`, `no-cache`)
			a.addCustomHeaders(w)
			http.ServeFile(w, r, a.directory)
			return
		}

		httperror.NotFound(w)
	})
}
