package env

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/ViBiOh/httputils/v2/pkg/request"
)

func TestNew(t *testing.T) {
	emptyString := ""
	envValue := "PATH,BASH,VERSION"

	var cases = []struct {
		intention string
		input     Config
		want      []string
	}{
		{
			"should work with empty values",
			Config{
				env: &emptyString,
			},
			nil,
		},
		{
			"should work with env value",
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

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := New(testCase.input).(app); !reflect.DeepEqual(result.keys, testCase.want) {
				t.Errorf("New(%#v) = %#v, want %#v", testCase.input, result.keys, testCase.want)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	user := os.Getenv("USER")
	os.Setenv("ESCAPE", `it's a "test"`)

	var cases = []struct {
		intention  string
		request    *http.Request
		env        string
		want       string
		wantStatus int
	}{
		{
			"should respond to OPTIONS request",
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"",
			"",
			http.StatusOK,
		},
		{
			"should reject non GET/OPTIONS request",
			httptest.NewRequest(http.MethodPost, "/", nil),
			"",
			"",
			http.StatusMethodNotAllowed,
		},
		{
			"should return empty JSON if no key",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			"{}",
			http.StatusOK,
		},
		{
			"should return asked keys",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"USER,ESCAPE",
			fmt.Sprintf(`{"ESCAPE":"it's a \"test\"","USER":"%s"}`, user),
			http.StatusOK,
		},
		{
			"should return empty value for not found keys",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"USER,UNKNOWN_ENV_VAR",
			fmt.Sprintf(`{"UNKNOWN_ENV_VAR":"","USER":"%s"}`, user),
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			a := New(Config{
				env: &testCase.env,
			})
			a.Handler().ServeHTTP(writer, testCase.request)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("Handler(%#v) = %d, want status %d", testCase.request, result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("Handler(%#v) = %s, want %s", testCase.request, string(result), testCase.want)
			}
		})
	}
}
