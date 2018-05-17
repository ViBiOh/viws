package env

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/httpjson"
	"github.com/ViBiOh/httputils/pkg/tools"
)

// App stores informations
type App struct {
	keys []string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	var keys []string

	env := strings.TrimSpace(*config[`env`])
	if env != `` {
		keys = strings.Split(env, `,`)
	}

	return &App{
		keys: keys,
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`env`: flag.String(tools.ToCamel(fmt.Sprintf(`%sEnv`, prefix)), ``, `[env] Environments key variables to expose, comma separated`),
	}
}

// Handler for net/http package returning environment variables in JSON
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		} else if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		env := make(map[string]string)

		for _, key := range a.keys {
			if value := os.Getenv(key); value != `` {
				env[key] = value
			} else {
				env[key] = ``
			}
		}

		if err := httpjson.ResponseJSON(w, http.StatusOK, env, httpjson.IsPretty(r.URL.RawQuery)); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}
