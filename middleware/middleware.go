package middleware

import "net/http"

type Interface interface {
	Wrap(http.Handler) http.Handler
}
