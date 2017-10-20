package writer

import (
	"bytes"
	"net/http"
)

// FakeResponseWriter implements http.ResponseWriter by storing content in memory
type FakeResponseWriter struct {
	status  int
	header  http.Header
	content *bytes.Buffer
}

// Content return current contant buffer
func (w *FakeResponseWriter) Content() *bytes.Buffer {
	return w.content
}

// Status return current writer status
func (w *FakeResponseWriter) Status() int {
	return w.status
}

// SetStatus set current writer status
func (w *FakeResponseWriter) SetStatus(status int) {
	w.status = status
}

// Header cf. https://golang.org/pkg/net/http/#ResponseWriter
func (w *FakeResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}

	return w.header
}

// Write cf. https://golang.org/pkg/net/http/#ResponseWriter
func (w *FakeResponseWriter) Write(content []byte) (int, error) {
	if w.content == nil {
		w.content = bytes.NewBuffer(make([]byte, 0, 1024))
	}

	return w.content.Write(content)
}

// WriteHeader cf. https://golang.org/pkg/net/http/#ResponseWriter
func (w *FakeResponseWriter) WriteHeader(status int) {
	w.status = status
}
