package pochi

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/lestrrat-go/pochi/middleware"
)

type PathSpec struct {
	pattern            string
	middlewares        []Middleware
	compiled           atomic.Bool
	handler            http.Handler
	inheritMiddlewares bool
}

func Path(pattern string) *PathSpec {
	return &PathSpec{
		pattern:            strings.TrimSuffix(pattern, "/"),
		inheritMiddlewares: true,
	}
}

func (p *PathSpec) Pattern() string {
	return p.pattern
}

func (p *PathSpec) Use(middlewares ...Middleware) *PathSpec {
	p.middlewares = append(p.middlewares, middlewares...)
	return p
}

func (p *PathSpec) InheritMiddlewares(v bool) *PathSpec {
	p.inheritMiddlewares = v
	return p
}

func (p *PathSpec) PrependMiddlewares(middlewares ...Middleware) *PathSpec {
	p.middlewares = append(middlewares, p.middlewares...)
	return p
}

func (p *PathSpec) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.compile()
	p.handler.ServeHTTP(w, req)
}

func (p *PathSpec) compile(ancestors ...*PathSpec) {
	if p.compiled.Load() {
		return
	}
	for _, m := range p.middlewares {
		p.handler = m.Wrap(p.handler)
	}
	p.compiled.Store(true)
}

func (p *PathSpec) Get(h http.Handler) *PathSpec {
	p.handler = middleware.RestrictMethod(http.MethodGet).Wrap(h)
	return p
}
