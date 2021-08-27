package env

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
)

// App of package
type App struct {
	keys []string
}

// Config of package
type Config struct {
	env *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		env: flags.New(prefix, "env", "Env").Default("", overrides).Label("Environments key variables to expose, comma separated").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	var keys []string

	if *config.env != "" {
		keys = strings.Split(*config.env, ",")
	}

	return App{
		keys: keys,
	}
}

// Handler for net/http package returning environment variables in JSON
func (a App) Handler() http.Handler {
	env := make(map[string]string)
	for _, key := range a.keys {
		env[key] = os.Getenv(key)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		httpjson.Write(w, http.StatusOK, env)
	})
}
