// Implements HTTP server handler that injects the LiveReload script tag into requested webpages.
//
// The code in this file is based on devserver code (https://github.com/prisoner/devserver).
//
package main

import (
	"bytes"
	"net/http"
	"strings"
)

type lrResponseWriter struct {
	http.ResponseWriter
	urlprefix string
}

func (w *lrResponseWriter) Write(bs []byte) (int, error) {
	const script = "<script src=\"http://localhost:35729/livereload.js\"></script>"
	if strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		bs = bytes.Replace(bs, []byte("</body>"), []byte(script+"\n</body>"), 1)
		if w.urlprefix != "" {
			bs = bytes.Replace(bs, []byte("href=\""+w.urlprefix), []byte("href=\""), -1)
			bs = bytes.Replace(bs, []byte("src=\""+w.urlprefix), []byte("src=\""), -1)
		}
	}
	return w.ResponseWriter.Write(bs)
}

func (w *lrResponseWriter) WriteHeader(code int) {
	if strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		h := w.Header()
		h.Del("Content-Length")
		h.Del("Last-Modified")
		h.Del("Date")
	}
	w.ResponseWriter.WriteHeader(code)
}

type lrHandler struct {
	http.Handler
	urlprefix string
}

func (h *lrHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w2 := &lrResponseWriter{
		ResponseWriter: w,
		urlprefix:      h.urlprefix,
	}
	h.Handler.ServeHTTP(w2, r)
}
