package env

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
)

// Config of package
type Config struct {
	env *string
}

// App of package
type App interface {
	Handler() http.Handler
}

type app struct {
	keys []string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		env: flags.New(prefix, "env").Name("Env").Default(flags.Default("Env", "", overrides)).Label("Environments key variables to expose, comma separated").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	var keys []string

	env := strings.TrimSpace(*config.env)
	if env != "" {
		keys = strings.Split(env, ",")
	}

	return app{
		keys: keys,
	}
}

// Handler for net/http package returning environment variables in JSON
func (a app) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		env := make(map[string]string)

		for _, key := range a.keys {
			env[key] = os.Getenv(key)
		}

		httpjson.Write(w, http.StatusOK, env, httpjson.IsPretty(r))
	})
}
