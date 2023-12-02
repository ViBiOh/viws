package env

import (
	"flag"
	"net/http"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
)

type Config struct {
	Env []string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Env", "Environments key variable to expose").Prefix(prefix).DocPrefix("env").StringSliceVar(fs, &config.Env, nil, overrides)

	return &config
}

type Service struct {
	keys []string
}

func New(config *Config) Service {
	return Service{
		keys: config.Env,
	}
}

func (s Service) Handler() http.Handler {
	env := make(map[string]string)
	for _, key := range s.keys {
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

		httpjson.Write(r.Context(), w, http.StatusOK, env)
	})
}
