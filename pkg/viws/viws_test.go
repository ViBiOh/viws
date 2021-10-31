package viws

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/request"
)

var exempleDir = "../../example/"

func TestFlags(t *testing.T) {
	cases := []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -directory string\n    \t[viws] Directory to serve {SIMPLE_DIRECTORY} (default \"/www/\")\n  -headers string\n    \t[viws] Custom headers, tilde separated (e.g. content-language:fr~X-UA-Compatible:test) {SIMPLE_HEADERS}\n  -spa\n    \t[viws] Indicate Single Page Application mode {SIMPLE_SPA}\n",
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
	exampleHeader := "= X-UA-Compatible:ie=edge~X-UA-Compatible:ie=edge~content-language:fr~invalidformat"

	cases := []struct {
		intention string
		input     Config
		want      App
	}{
		{
			"minimal config",
			Config{
				directory: &exempleDir,
				headers:   &emptyString,
				spa:       &falseVar,
			},
			App{
				spa:       false,
				directory: exempleDir,
				headers:   http.Header{},
			},
		},
		{
			"spa config",
			Config{
				directory: &exempleDir,
				headers:   &emptyString,
				spa:       &trueVar,
			},
			App{
				spa:       true,
				directory: exempleDir,
				headers:   http.Header{},
			},
		},
		{
			"headers",
			Config{
				directory: &exempleDir,
				headers:   &exampleHeader,
				spa:       &falseVar,
			},
			App{
				spa:       false,
				directory: exempleDir,
				headers: http.Header{
					"X-Ua-Compatible":  []string{"ie=edge"},
					"Content-Language": []string{"fr"},
				},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := New(testCase.input); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("New() = %+v, want %+v", result, testCase.want)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	cases := []struct {
		intention  string
		app        App
		request    *http.Request
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		{
			"invalid method",
			App{},
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"",
			http.StatusMethodNotAllowed,
			nil,
		},
		{
			"head index",
			App{
				directory: exempleDir,
			},
			httptest.NewRequest(http.MethodHead, "/", nil),
			"",
			http.StatusNoContent,
			nil,
		},
		{
			"path with dots",
			App{
				directory: exempleDir,
			},
			httptest.NewRequest(http.MethodHead, "/../index.html", nil),
			"path with dots are not allowed\n",
			http.StatusBadRequest,
			nil,
		},
		{
			"get index",
			App{
				directory: exempleDir,
			},
			httptest.NewRequest(http.MethodGet, "/", nil),
			`<!DOCTYPE HTML>
<html lang="en">
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
				cacheControlHeader: {noCacheValue},
			},
		},
		{
			"get file with header",
			App{
				directory: exempleDir,
				headers: http.Header{
					"Etag": []string{"test"},
				},
			},
			httptest.NewRequest(http.MethodGet, "/index.js", nil),
			`console.log('Ready');
`,
			http.StatusOK,
			http.Header{
				"Etag":             []string{"test"},
				cacheControlHeader: {"public, max-age=864000"},
			},
		},
		{
			"head not found",
			App{
				directory: exempleDir,
			},
			httptest.NewRequest(http.MethodHead, "/404.html", nil),
			"",
			http.StatusNotFound,
			http.Header{
				cacheControlHeader: {""},
			},
		},
		{
			"get not found",
			App{
				directory: exempleDir,
			},
			httptest.NewRequest(http.MethodGet, "/404.html", nil),
			`¯\_(ツ)_/¯
`,
			http.StatusNotFound,
			http.Header{
				cacheControlHeader: {noCacheValue},
			},
		},
		{
			"get not found with file",
			App{
				directory: "../../example/404/",
			},
			httptest.NewRequest(http.MethodGet, "/nowhere", nil),
			`<!DOCTYPE HTML>
<html lang="en">
  <head>
    <title>viws</title>
  </head>
  <body>
    <h1>Not found!</h1>
  </body>
</html>
`,
			http.StatusNotFound,
			http.Header{
				cacheControlHeader: {noCacheValue},
			},
		},
		{
			"get not found with spa",
			App{
				directory: exempleDir,
				spa:       true,
			},
			httptest.NewRequest(http.MethodGet, "/user/1234", nil),
			`<!DOCTYPE HTML>
<html lang="en">
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
				cacheControlHeader: {noCacheValue},
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

type discardResponseWriter struct {
	h http.Header
}

func newDiscardResponseWriter() discardResponseWriter {
	return discardResponseWriter{
		h: http.Header{},
	}
}

func (drw discardResponseWriter) Header() http.Header {
	return drw.h
}

func (drw discardResponseWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (drw discardResponseWriter) WriteHeader(_ int) {
}

func BenchmarkServeFile(b *testing.B) {
	headers := http.Header{}
	headers.Add("X-UA-Compatible", "ie=edge")
	headers.Add("content-language", "fr")

	instance := App{
		directory: exempleDir,
		headers:   headers,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := newDiscardResponseWriter()
	now := time.Now()

	for i := 0; i < b.N; i++ {
		instance.serveFile(recorder, req, "../../example/404/index.html", now)
	}
}
