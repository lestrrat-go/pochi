package middleware

import "net/http"

// RestrictMethod returns a new middleware that restricts the HTTP method
// to the handler(s) downstream. If the method does not match, it will return
// a 405 Method Not Allowed error
func RestrictMethod(method string) Interface {
	return &restrictMethodBuilder{method: method}
}

type restrictMethodBuilder struct {
	method string
}

func (m *restrictMethodBuilder) Wrap(h http.Handler) http.Handler {
	return restrictMethod{next: h, method: m.method}
}

type restrictMethod struct {
	next   http.Handler
	method string
}

func (m restrictMethod) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != m.method {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	m.next.ServeHTTP(w, r)
}
