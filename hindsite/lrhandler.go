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
	newbody = []byte("<script src=\"http://127.0.0.1:35729/livereload.js\"></script>\n</body>")
)

type ssResponseWriter struct {
	http.ResponseWriter
}

func (w *ssResponseWriter) Write(bs []byte) (int, error) {
	if strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		bs = bytes.Replace(bs, body, newbody, -1)
	}
	return w.ResponseWriter.Write(bs)
}

func (w *ssResponseWriter) WriteHeader(code int) {
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

type ssHandler struct {
	http.Handler
}

func (h *ssHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.String() == "/favicon.ico" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w2 := &ssResponseWriter{
		ResponseWriter: w,
	}
	h.Handler.ServeHTTP(w2, r)
}
