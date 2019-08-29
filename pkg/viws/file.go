package viws

import (
	"os"
	"path"

	"github.com/ViBiOh/httputils/pkg/errors"
)

const (
	indexFilename = "index.html"
)

func getFileToServe(parts ...string) (string, error) {
	path := path.Join(parts...)

	info, err := os.Stat(path)
	if err != nil {
		return path, errors.WithStack(err)
	}

	if !info.IsDir() {
		return path, nil
	}

	return getFileToServe(append(parts, indexFilename)...)
}
