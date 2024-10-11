package pochi

import "net/http"

type Middleware interface {
	Wrap(http.Handler) http.Handler
}
