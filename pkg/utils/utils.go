package utils

import (
	"os"
	"path"
)

const (
	indexFilename = "index.html"
)

// IsFileExist check if concatenated paths are available for serving
func IsFileExist(parts ...string) *string {
	fullPath := path.Join(parts...)
	info, err := os.Stat(fullPath)

	if err != nil {
		return nil
	}

	if info.IsDir() {
		if IsFileExist(append(parts, indexFilename)...) == nil {
			return nil
		}
	}

	return &fullPath
}
