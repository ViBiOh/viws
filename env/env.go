package env

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/httputils/httperror"
	"github.com/ViBiOh/httputils/json"
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

		if err := json.ResponseJSON(w, http.StatusOK, env, json.IsPretty(r.URL.RawQuery)); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}
