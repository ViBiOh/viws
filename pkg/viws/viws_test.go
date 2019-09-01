package viws

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/pkg/request"
)

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

			if result, _ := request.ReadBody(writer.Result().Body); string(result) != testCase.want {
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
