package viws

import (
	"os"
	"path"
)

func getFileToServe(parts ...string) (string, error) {
	filepath := path.Join(parts...)

	info, err := os.Stat(filepath)
	if err != nil {
		return filepath, err
	}

	if !info.IsDir() {
		return filepath, nil
	}

	return getFileToServe(append(parts, "index.html")...)
}
