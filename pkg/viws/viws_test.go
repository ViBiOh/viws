package viws

import (
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/request"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -directory string\n    \t[viws] Directory to serve {SIMPLE_DIRECTORY} (default \"/www/\")\n  -headers string\n    \t[viws] Custom headers, tilde separated (e.g. content-language:fr~X-UA-Compatible:test) {SIMPLE_HEADERS}\n  -push string\n    \t[viws] Paths for HTTP/2 Server Push on index, comma separated {SIMPLE_PUSH}\n  -spa\n    \t[viws] Indicate Single Page Application mode {SIMPLE_SPA}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	falseVar := false
	trueVar := true
	emptyString := ""
	exempleDir := "../../example/"
	examplePush := "index.js,index.css"
	exampleHeader := "X-UA-Compatible:ie=edge~content-language:fr~invalidformat"

	var cases = []struct {
		intention string
		input     Config
		want      App
		wantErr   error
	}{
		{
			"minimal config",
			Config{
				directory: &exempleDir,
				headers:   &emptyString,
				spa:       &falseVar,
				push:      &emptyString,
			},
			app{
				spa:       false,
				directory: "../../example/",
			},
			nil,
		},
		{
			"spa config",
			Config{
				directory: &exempleDir,
				headers:   &emptyString,
				spa:       &trueVar,
				push:      &emptyString,
			},
			app{
				spa:       true,
				directory: "../../example/",
			},
			nil,
		},
		{
			"empty config",
			Config{
				directory: &emptyString,
				headers:   &emptyString,
				spa:       &falseVar,
				push:      &emptyString,
			},
			nil,
			errors.New("directory  is unreachable or does not contains index: stat : no such file or directory"),
		},
		{
			"pushPaths",
			Config{
				directory: &exempleDir,
				headers:   &emptyString,
				spa:       &falseVar,
				push:      &examplePush,
			},
			app{
				spa:       false,
				directory: "../../example/",
				pushPaths: []string{
					"index.js",
					"index.css",
				},
			},
			nil,
		},
		{
			"header",
			Config{
				directory: &exempleDir,
				headers:   &exampleHeader,
				spa:       &falseVar,
				push:      &emptyString,
			},
			app{
				spa:       false,
				directory: "../../example/",
				headers: map[string]string{
					"X-UA-Compatible":  "ie=edge",
					"content-language": "fr",
				},
			},
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := New(testCase.input)

			failed := false

			if err == nil && testCase.wantErr != nil {
				failed = true
			} else if err != nil && testCase.wantErr == nil {
				failed = true
			} else if err != nil && err.Error() != testCase.wantErr.Error() {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("New() = (%#v, `%s`), want (%#v, `%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	var cases = []struct {
		intention  string
		app        app
		request    *http.Request
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		{
			"invalid method",
			app{},
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"",
			http.StatusMethodNotAllowed,
			nil,
		},
		{
			"head index",
			app{
				directory: "../../example/",
			},
			httptest.NewRequest(http.MethodHead, "/", nil),
			"",
			http.StatusNoContent,
			nil,
		},
		{
			"get index",
			app{
				directory: "../../example/",
			},
			httptest.NewRequest(http.MethodGet, "/", nil),
			`<!DOCTYPE HTML>
<html>
  <head>
    <title>viws</title>
    <link rel="stylesheet" href="/index.css">
    <script src="/index.js"></script>
  </head>
  <body>
    <h1>Hello World!</h1>
  </body>
</html>
`,
			http.StatusOK,
			nil,
		},
		{
			"get file with header",
			app{
				directory: "../../example/",
				headers: map[string]string{
					"Etag": "test",
				},
			},
			httptest.NewRequest(http.MethodGet, "/index.js", nil),
			`console.log('Ready');
`,
			http.StatusOK,
			http.Header{
				"Etag": []string{"test"},
			},
		},
		{
			"head not found",
			app{
				directory: "../../example/",
			},
			httptest.NewRequest(http.MethodHead, "/404.html", nil),
			"",
			http.StatusNotFound,
			nil,
		},
		{
			"get not found",
			app{
				directory: "../../example/",
			},
			httptest.NewRequest(http.MethodGet, "/404.html", nil),
			`¯\_(ツ)_/¯
`,
			http.StatusNotFound,
			nil,
		},
		{
			"get not found with file",
			app{
				directory:    "../../example/",
				notFoundPath: "../../example/index.js",
			},
			httptest.NewRequest(http.MethodGet, "/404.html", nil),
			`console.log('Ready');
`,
			http.StatusNotFound,
			nil,
		},
		{
			"get not found with spa",
			app{
				directory: "../../example/",
				spa:       true,
			},
			httptest.NewRequest(http.MethodGet, "/user/1234", nil),
			`<!DOCTYPE HTML>
<html>
  <head>
    <title>viws</title>
    <link rel="stylesheet" href="/index.css">
    <script src="/index.js"></script>
  </head>
  <body>
    <h1>Hello World!</h1>
  </body>
</html>
`,
			http.StatusOK,
			http.Header{
				"Cache-Control": {"no-cache"},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			testCase.app.Handler().ServeHTTP(writer, testCase.request)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("Status %d, want %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("Body `%s`, want `%s`", string(result), testCase.want)
			}

			for key := range testCase.wantHeader {
				want := testCase.wantHeader.Get(key)
				if result := writer.Header().Get(key); result != want {
					t.Errorf("%s Header = `%s`, want `%s`", key, result, want)
				}
			}
		})
	}
}
