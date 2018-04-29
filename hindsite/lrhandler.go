// Implements HTTP server handler that injects the LiveReload script tag into requested webpages.
//
// The code in this file is based on devserver code (https://github.com/prisoner/devserver).
//
package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var (
	body    = []byte("</body>")
	newbody = []byte("<script src=\"http://localhost:35729/livereload.js\"></script>\n</body>")
)

type lrResponseWriter struct {
	http.ResponseWriter
}

func (w *lrResponseWriter) Write(bs []byte) (int, error) {
	if strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		bs = bytes.Replace(bs, body, newbody, 1)
	}
	return w.ResponseWriter.Write(bs)
}

func (w *lrResponseWriter) WriteHeader(code int) {
	if strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		l, _ := strconv.Atoi(w.Header().Get("Content-Length"))
		l = l + len(newbody) - len(body)
		h := w.Header()
		h.Set("Content-Length", fmt.Sprintf("%d", l))
		h.Del("Last-Modified")
		h.Del("Date")
	}
	w.ResponseWriter.WriteHeader(code)
}

type lrHandler struct {
	http.Handler
}

func (h *lrHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.String() == "/favicon.ico" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w2 := &lrResponseWriter{
		ResponseWriter: w,
	}
	h.Handler.ServeHTTP(w2, r)
}
