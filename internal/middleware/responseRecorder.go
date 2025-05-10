package middleware

import (
	"bytes"
	"net/http"
)

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{w, http.StatusOK, bytes.NewBuffer(nil)}
}

func (ww *ResponseRecorder) WriteHeader(code int) {
	ww.StatusCode = code
	ww.ResponseWriter.WriteHeader(code)
}

func (ww *ResponseRecorder) Write(resp []byte) (int, error) {
	ww.Body.Write(resp)
	return ww.ResponseWriter.Write(resp)
}
