package viws

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func (a App) serveNotFound(w http.ResponseWriter) {
	notFoundPath, _, err := getFileToServe(a.directory, notFoundFilename)
	if err != nil {
		httperror.NotFound(w)
		return
	}

	a.serve(w, http.StatusNotFound, notFoundPath)
}

func (a App) serve(w http.ResponseWriter, status int, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			logger.WithField("dir", a.directory).Error("unable to close file: %s", err)
		}
	}()

	contentType := mime.TypeByExtension(filename)
	if len(contentType) == 0 {
		contentType = "text/html; charset=utf-8"
	}

	w.Header().Add("Content-Type", contentType)
	w.Header().Add(cacheControlHeader, noCacheValue)
	w.WriteHeader(status)

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	if _, err = io.CopyBuffer(w, file, buffer.Bytes()); err != nil {
		logger.WithField("dir", a.directory).Error("unable to copy content to writer: %s", err)
	}
}
