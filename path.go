package pochi

import (
	"net/http"
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

// Path creates a new PathSpec object with the given pattern.
func Path(pattern string) *PathSpec {
	return &PathSpec{
		pattern:            pattern,
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

func (p *PathSpec) compile() {
	if p.compiled.Load() {
		return
	}
	for _, m := range p.middlewares {
		p.handler = m.Wrap(p.handler)
	}
	p.compiled.Store(true)
}

func (p *PathSpec) Method(m string, h http.Handler) *PathSpec {
	p.handler = middleware.RestrictMethod(m).Wrap(h)
	return p
}

func (p *PathSpec) Get(h http.Handler) *PathSpec {
	return p.Method(http.MethodGet, h)
}

func (p *PathSpec) Head(h http.Handler) *PathSpec {
	return p.Method(http.MethodHead, h)
}

func (p *PathSpec) Post(h http.Handler) *PathSpec {
	return p.Method(http.MethodPost, h)
}

func (p *PathSpec) Put(h http.Handler) *PathSpec {
	return p.Method(http.MethodPut, h)
}

func (p *PathSpec) Patch(h http.Handler) *PathSpec {
	return p.Method(http.MethodPatch, h)
}

func (p *PathSpec) Delete(h http.Handler) *PathSpec {
	return p.Method(http.MethodDelete, h)
}

func (p *PathSpec) Connect(h http.Handler) *PathSpec {
	return p.Method(http.MethodConnect, h)
}

func (p *PathSpec) Options(h http.Handler) *PathSpec {
	return p.Method(http.MethodOptions, h)
}

func (p *PathSpec) Trace(h http.Handler) *PathSpec {
	return p.Method(http.MethodTrace, h)
}
