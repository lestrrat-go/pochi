package pochi

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/lestrrat-go/pochi/middleware"
)

type PathSpec struct {
	pattern string
	// list of middlewares that were actually applied to this path
	directMiddlewares []Middleware
	// list of middlewares that were inherited from the parent path
	inheritedMiddlewares []Middleware

	// set to true if the object's handler has been "compiled" with all of
	// the appropriate middlewares
	compiled           atomic.Bool
	rawHandler         http.Handler
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

func MountPath(prefix string, path *PathSpec) *PathSpec {
	prefix = strings.TrimSuffix(prefix, "/")
	p := Path(prefix + "/" + strings.TrimPrefix(path.pattern, "/"))
	p.rawHandler = path.rawHandler
	p.inheritMiddlewares = path.inheritMiddlewares
	if l := len(path.directMiddlewares); l > 0 {
		p.directMiddlewares = make([]Middleware, l)
		copy(p.directMiddlewares, path.directMiddlewares)
	}
	return p
}

func (p *PathSpec) HasHandler() bool {
	return p.rawHandler != nil
}

func (p *PathSpec) Pattern() string {
	return p.pattern
}

func (p *PathSpec) Use(middlewares ...Middleware) *PathSpec {
	p.directMiddlewares = append(p.directMiddlewares, middlewares...)
	return p
}

// Inherit specifies whether the middlewares from the parent path should be inherited.
func (p *PathSpec) Inherit(v bool) *PathSpec {
	p.inheritMiddlewares = v
	return p
}

func (p *PathSpec) InheritMiddlewares(middlewares ...Middleware) *PathSpec {
	p.inheritedMiddlewares = append(p.inheritedMiddlewares, middlewares...)
	return p
}

func (p *PathSpec) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.compile()
	p.handler.ServeHTTP(w, req)
}

func (p *PathSpec) invalidate() {
	p.compiled.Store(false)
	if len(p.inheritedMiddlewares) > 0 {
		p.inheritedMiddlewares = p.inheritedMiddlewares[:0]
	}
}

func (p *PathSpec) compile() {
	if p.compiled.Load() {
		return
	}

	p.handler = p.rawHandler
	if l := len(p.directMiddlewares); l > 0 {
		for i := l - 1; i >= 0; i-- {
			p.handler = p.directMiddlewares[i].Wrap(p.handler)
		}
	}

	if l := len(p.inheritedMiddlewares); l > 0 {
		for i := l - 1; i >= 0; i-- {
			p.handler = p.inheritedMiddlewares[i].Wrap(p.handler)
		}
	}
	p.compiled.Store(true)
}

func (p *PathSpec) Method(m string, h http.Handler) *PathSpec {
	p.rawHandler = middleware.RestrictMethod(m).Wrap(h)
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
