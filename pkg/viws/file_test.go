package viws

import (
	"errors"
	"testing"
)

func TestGetFileToServe(t *testing.T) {
	type args struct {
		directory string
		path      string
	}
	var cases = []struct {
		intention string
		args      args
		want      string
		wantErr   error
	}{
		{
			"empty call",
			args{},
			"",
			errors.New("stat : no such file or directory"),
		},
		{
			"local file",
			args{
				path: "file.go",
			},
			"file.go",
			nil,
		},
		{
			"concatenared dir path",
			args{
				directory: "../..",
				path:      "example",
			},
			"../../example/index.html",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			result, err := getFileToServe(tc.args.directory, tc.args.path)

			failed := false

			if err == nil && tc.wantErr != nil {
				failed = true
			} else if err != nil && tc.wantErr == nil {
				failed = true
			} else if err != nil && err.Error() != tc.wantErr.Error() {
				failed = true
			} else if result != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("getFileToServe() = (`%s`, `%s`), want (`%s`, `%s`)", result, err, tc.want, tc.wantErr)
			}
		})
	}
}
