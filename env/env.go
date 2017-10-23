package env

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/httputils"
)

var (
	keys    = flag.String(`env`, ``, `Environments key variables to expose, comma separated`)
	envKeys []string
)

// Init package
func Init() error {
	if *keys != `` {
		envKeys = strings.Split(*keys, `,`)
	}

	return nil
}

// Handler for net/http package returning environment variables in JSON
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := make(map[string]string)

		for _, key := range envKeys {
			if value := os.Getenv(key); value != `` {
				env[key] = value
			}
		}

		httputils.ResponseJSON(w, http.StatusOK, env, httputils.IsPretty(r.URL.RawQuery))
	})
}
