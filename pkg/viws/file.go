package viws

import (
	"os"
	"path/filepath"
	"time"
)

func getFileToServe(parts ...string) (string, time.Time, error) {
	filename := filepath.Join(parts...)

	info, err := os.Stat(filename)
	if err != nil {
		return "", time.Time{}, err
	}

	if !info.IsDir() {
		return filename, info.ModTime(), nil
	}

	return getFileToServe(filename, indexFilename)
}
