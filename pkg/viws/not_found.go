package viws

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
)

func (a App) serveNotFound(ctx context.Context, w http.ResponseWriter) {
	notFoundPath, _, err := getFileToServe(a.directory, notFoundFilename)
	if err != nil {
		httperror.NotFound(ctx, w)
		return
	}

	a.serve(ctx, w, http.StatusNotFound, notFoundPath)
}

func (a App) serve(ctx context.Context, w http.ResponseWriter, status int, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		httperror.InternalServerError(ctx, w, err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "close file", slog.String("dir", a.directory), slog.Any("error", err))
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
		slog.LogAttrs(ctx, slog.LevelError, "copy content to writer", slog.String("dir", a.directory), slog.Any("error", err))
	}
}
