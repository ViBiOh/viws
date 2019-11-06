package viws

import (
	"errors"
	"testing"
)

func TestGetFileToServe(t *testing.T) {
	var cases = []struct {
		intention string
		input     []string
		want      string
		wantErr   error
	}{
		{
			"nil call",
			nil,
			"",
			errors.New("stat : no such file or directory"),
		},
		{
			"local file",
			[]string{"file.go"},
			"file.go",
			nil,
		},
		{
			"concatenared dir path",
			[]string{"..", "..", "example"},
			"../../example/index.html",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := getFileToServe(testCase.input...)

			failed := false

			if err == nil && testCase.wantErr != nil {
				failed = true
			} else if err != nil && testCase.wantErr == nil {
				failed = true
			} else if err != nil && err.Error() != testCase.wantErr.Error() {
				failed = true
			} else if result != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("getFileToServe() = (%s, %s), want (%s, %s)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}
