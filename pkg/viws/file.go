package viws

import (
	"os"
	"path/filepath"
	"time"
)

func getFileToServe(parts ...string) (string, time.Time, error) {
	filepath := filepath.Join(parts...)

	info, err := os.Stat(filepath)
	if err != nil {
		return "", time.Time{}, err
	}

	if !info.IsDir() {
		return filepath, info.ModTime(), nil
	}

	return getFileToServe(filepath, indexFilename)
}
