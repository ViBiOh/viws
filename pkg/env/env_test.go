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

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
		wantType  string
	}{
		{
			`should add string env param to flags`,
			`env`,
			`*string`,
		},
	}

	for _, testCase := range cases {
		result := Flags(testCase.intention)[testCase.want]

		if result == nil {
			t.Errorf("%s\nFlags() = %+v, want `%s`", testCase.intention, result, testCase.want)
		}

		if fmt.Sprintf(`%T`, result) != testCase.wantType {
			t.Errorf("%s\nFlags() = `%T`, want `%s`", testCase.intention, result, testCase.wantType)
		}
	}
}

func Test_NewApp(t *testing.T) {
	emptyString := ``
	envValue := `PATH,BASH,VERSION`

	var cases = []struct {
		intention string
		input     map[string]*string
		want      []string
	}{
		{
			`should work with empty values`,
			map[string]*string{
				`env`: &emptyString,
			},
			nil,
		},
		{
			`should work with env value`,
			map[string]*string{
				`env`: &envValue,
			},
			[]string{
				`PATH`,
				`BASH`,
				`VERSION`,
			},
		},
	}

	for _, testCase := range cases {
		if result := NewApp(testCase.input); !reflect.DeepEqual(result.keys, testCase.want) {
			t.Errorf("%s\nNewApp(%+v) = %+v, want %+v", testCase.intention, testCase.input, result.keys, testCase.want)
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

		a := NewApp(map[string]*string{
			`env`: &testCase.env,
		})
		a.Handler().ServeHTTP(writer, testCase.request)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%s\nHandler(%+v) = %d, want status %d", testCase.intention, testCase.request, result, testCase.wantStatus)
		}

		if result, _ := request.ReadBody(writer.Result().Body); string(result) != testCase.want {
			t.Errorf("%s\nHandler(%+v) = %s, want %s", testCase.intention, testCase.request, string(result), testCase.want)
		}
	}
}
