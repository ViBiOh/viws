package env

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/ViBiOh/httputils/pkg/request"
)

func Test_New(t *testing.T) {
	emptyString := ``
	envValue := `PATH,BASH,VERSION`

	var cases = []struct {
		intention string
		input     Config
		want      []string
	}{
		{
			`should work with empty values`,
			Config{
				env: &emptyString,
			},
			nil,
		},
		{
			`should work with env value`,
			Config{
				env: &envValue,
			},
			[]string{
				`PATH`,
				`BASH`,
				`VERSION`,
			},
		},
	}

	for _, testCase := range cases {
		if result := New(testCase.input); !reflect.DeepEqual(result.keys, testCase.want) {
			t.Errorf("%s\nNew(%+v) = %+v, want %+v", testCase.intention, testCase.input, result.keys, testCase.want)
		}
	}
}

func Test_Handler(t *testing.T) {
	user := os.Getenv(`USER`)
	path := os.Getenv(`PATH`)

	var cases = []struct {
		intention  string
		request    *http.Request
		env        string
		want       string
		wantStatus int
	}{
		{
			`should respond to OPTIONS request`,
			httptest.NewRequest(http.MethodOptions, `/`, nil),
			``,
			``,
			http.StatusOK,
		},
		{
			`should reject non GET/OPTIONS request`,
			httptest.NewRequest(http.MethodPost, `/`, nil),
			``,
			``,
			http.StatusMethodNotAllowed,
		},
		{
			`should return empty JSON if no key`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			``,
			`{}`,
			http.StatusOK,
		},
		{
			`should return asked keys`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			`USER,PATH`,
			fmt.Sprintf(`{"PATH":"%s","USER":"%s"}`, path, user),
			http.StatusOK,
		},
		{
			`should return empty value for not found keys`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			`USER,UNKNOWN_ENV_VAR`,
			fmt.Sprintf(`{"UNKNOWN_ENV_VAR":"","USER":"%s"}`, user),
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()

		a := New(Config{
			env: &testCase.env,
		})
		a.Handler().ServeHTTP(writer, testCase.request)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%s\nHandler(%+v) = %d, want status %d", testCase.intention, testCase.request, result, testCase.wantStatus)
		}

		if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
			t.Errorf("%s\nHandler(%+v) = %s, want %s", testCase.intention, testCase.request, string(result), testCase.want)
		}
	}
}
