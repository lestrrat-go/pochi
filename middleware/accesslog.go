package middleware

import (
	"fmt"
	"net/http"
)

type Interface interface {
	Wrap(http.Handler) http.Handler
}

func AccessLog() Interface {
	return &accessLogBuilder{}
}

type accessLogBuilder struct{}

func (m *accessLogBuilder) Wrap(h http.Handler) http.Handler {
	return accessLog{next: h}
}

type accessLog struct {
	next http.Handler
}

func (m accessLog) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Do something with the request
	fmt.Println("HTTP GET", r.URL.Path)
	m.next.ServeHTTP(w, r)
}
