package viws

import (
	"os"
	"path"
)

func getFileToServe(parts ...string) (string, error) {
	path := path.Join(parts...)

	info, err := os.Stat(path)
	if err != nil {
		return path, err
	}

	if !info.IsDir() {
		return path, nil
	}

	return getFileToServe(append(parts, "index.html")...)
}
