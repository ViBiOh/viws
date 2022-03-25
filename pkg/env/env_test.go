package env

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func TestFlags(t *testing.T) {
	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -env string\n    \t[env] Environments key variables to expose, comma separated {SIMPLE_ENV}\n",
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			fs := flag.NewFlagSet(intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != tc.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, tc.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	emptyString := ""
	envValue := "PATH,BASH,VERSION"

	cases := map[string]struct {
		input Config
		want  []string
	}{
		"should work with empty values": {
			Config{
				env: &emptyString,
			},
			nil,
		},
		"should work with env value": {
			Config{
				env: &envValue,
			},
			[]string{
				"PATH",
				"BASH",
				"VERSION",
			},
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if result := New(tc.input); !reflect.DeepEqual(result.keys, tc.want) {
				t.Errorf("New() = %+v, want %+v", result.keys, tc.want)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	user := os.Getenv("USER")
	os.Setenv("ESCAPE", `it's a "test"`)

	cases := map[string]struct {
		request    *http.Request
		env        string
		want       string
		wantStatus int
	}{
		"should respond to OPTIONS request": {
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"",
			"",
			http.StatusOK,
		},
		"should reject non GET/OPTIONS request": {
			httptest.NewRequest(http.MethodPost, "/", nil),
			"",
			"",
			http.StatusMethodNotAllowed,
		},
		"should return empty JSON if no key": {
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			"{}\n",
			http.StatusOK,
		},
		"should return asked keys": {
			httptest.NewRequest(http.MethodGet, "/", nil),
			"USER,ESCAPE",
			fmt.Sprintf("{\"ESCAPE\":\"it's a \\\"test\\\"\",\"USER\":\"%s\"}\n", user),
			http.StatusOK,
		},
		"should return empty value for not found keys": {
			httptest.NewRequest(http.MethodGet, "/", nil),
			"USER,UNKNOWN_ENV_VAR",
			fmt.Sprintf("{\"UNKNOWN_ENV_VAR\":\"\",\"USER\":\"%s\"}\n", user),
			http.StatusOK,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			a := New(Config{
				env: &tc.env,
			})
			a.Handler().ServeHTTP(writer, tc.request)

			if result := writer.Code; result != tc.wantStatus {
				t.Errorf("Handler() = %d, want status %d", result, tc.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != tc.want {
				t.Errorf("Handler() = `%s`, want `%s`", string(result), tc.want)
			}
		})
	}
}
