package pochi

import (
	"errors"
	"fmt"
	"iter"
	"net/http"
	"path"
	"strings"

	"github.com/lestrrat-go/trie/v2"
)

type Router interface {
	http.Handler
	MatchRoute(string) (*PathSpec, bool)
	Route(...*PathSpec) error
	Walk(RouteVisitor)
}

type router struct {
	paths       *pathtrie
	cachedPaths map[string]*PathSpec
}

func NewRouter() Router {
	return &router{
		paths:       newPathtrie(),
		cachedPaths: make(map[string]*PathSpec),
	}
}

type RouteVisitor interface {
	Visit(string, *PathSpec)
}

type RouteVisitFunc func(string, *PathSpec)

func (f RouteVisitFunc) Visit(s string, spec *PathSpec) {
	f(s, spec)
}

func iterToPath(nodes iter.Seq[trie.Node[string, *PathSpec]]) string {
	var buf strings.Builder
	count := 0
	for n := range nodes {
		if count > 0 {
			buf.WriteByte('/')
		}
		buf.WriteString(n.Key())
		count++
	}
	return buf.String()
}

func (r *router) Walk(v RouteVisitor) {
	trie.Walk(r.paths.impl, trie.VisitFunc[string, *PathSpec](func(n trie.Node[string, *PathSpec], _ trie.VisitMetadata) bool {
		if spec := n.Value(); spec != nil {
			v.Visit(iterToPath(n.Ancestors())+"/"+n.Key(), spec)
		}
		return true
	}))
}

var errInvalidPath = errors.New("invalid path")

func ErrInvalidPath() error {
	return errInvalidPath
}

func (r *router) Route(specs ...*PathSpec) error {
	for _, spec := range specs {
		if !strings.HasPrefix(spec.pattern, "/") {
			return fmt.Errorf("paths must be absolute %q: %w", spec.pattern, ErrInvalidPath())
		}

		r.paths.Put(spec.pattern, spec)
		node, ok := r.paths.GetNode(spec.pattern)
		if !ok {
			panic("failed to fetch node that we just inserted")
		}
		if spec.inheritMiddlewares {
			for ancestor := range node.Ancestors() {
				rootPath := ancestor.First()
				if rootPath == node {
					continue
				}
				if rootPath.Key() != "" {
					continue
				}
				ancestorSpec := rootPath.Value()
				if ancestorSpec != nil {
					spec.PrependMiddlewares(ancestorSpec.middlewares...)
				}
			}
		}
	}
	return nil
}

func (r *router) MatchRoute(p string) (*PathSpec, bool) {
	for p != "" {
		spec, ok := r.paths.Get(p)
		if !ok && p != "/" {
			if p[len(p)-1] == '/' {
				// if the path ends with a '/' (and is not root)
				// strip it and try again
				p = p[:len(p)-1]
			}

			p = path.Dir(p)
			if p[len(p)-1] != '/' {
				p += "/"
			}
			continue
		}
		if spec == nil {
			return nil, false
		}
		return spec, true
	}
	return nil, false
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pathkey := req.URL.Path //strings.TrimSuffix(req.URL.Path, "/")
	spec, ok := r.cachedPaths[pathkey]
	if !ok {
		spec, ok = r.MatchRoute(pathkey)
		if !ok || spec.handler == nil {
			http.NotFound(w, req)
			return
		}
	}
	r.cachedPaths[pathkey] = spec
	spec.ServeHTTP(w, req)
}
