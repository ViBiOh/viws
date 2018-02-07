package env

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
)

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`env`: flag.String(tools.ToCamel(prefix+`Env`), ``, `[env] Environments key variables to expose, comma separated`),
	}
}

// Handler for net/http package returning environment variables in JSON
func Handler(config map[string]*string) http.Handler {
	envKeys := strings.Split(*config[`env`], `,`)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		env := make(map[string]string)

		for _, key := range envKeys {
			if value := os.Getenv(key); value != `` {
				env[key] = value
			}
		}

		if err := httputils.ResponseJSON(w, http.StatusOK, env, httputils.IsPretty(r.URL.RawQuery)); err != nil {
			httputils.InternalServerError(w, err)
		}
	})
}
