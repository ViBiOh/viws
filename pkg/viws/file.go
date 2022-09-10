package viws

import (
	"os"
	"path/filepath"
)

func getFileToServe(parts ...string) (string, os.FileInfo, error) {
	filename := filepath.Join(parts...)

	info, err := os.Stat(filename)
	if err != nil {
		return "", nil, err
	}

	if !info.IsDir() {
		return filename, info, nil
	}

	return getFileToServe(filename, indexFilename)
}
