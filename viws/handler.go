package viws

import (
	"log"
	"net/http"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/viws/utils"
)

// ServerPushHandler add server push when serving index
func ServerPushHandler(next http.Handler, pushPaths []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path == `/` && len(pushPaths) > 0 {
			if pusher, ok := w.(http.Pusher); ok {
				for _, path := range pushPaths {
					if err := pusher.Push(path, nil); err != nil {
						log.Printf(`Failed to push %s: %v`, path, err)
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// FileHandler serve file given configuration
func FileHandler(directory string, spa bool, notFound bool, notFoundPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if filename := utils.IsFileExist(directory, r.URL.Path); filename != nil {
			http.ServeFile(w, r, *filename)
		} else if notFound {
			w.WriteHeader(http.StatusNotFound)
			http.ServeFile(w, r, notFoundPath)
		} else if spa {
			w.Header().Add(`Cache-Control`, `no-cache`)
			http.ServeFile(w, r, directory)
		} else {
			httputils.NotFound(w)
		}
	})
}
