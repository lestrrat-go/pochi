package pochi

import (
	"iter"
	"net/http"
	"strings"

	"github.com/lestrrat-go/trie/v2"
)

type Router interface {
	http.Handler
	Route(...*PathSpec)
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

func (r *router) Route(specs ...*PathSpec) {
	for _, spec := range specs {
		r.paths.Put(spec.pattern, spec)
		node, ok := r.paths.GetNode(spec.pattern)
		if !ok {
			panic("failed to fetch node that we just inserted")
		}
		if spec.inheritMiddlewares {
			for ancestor := range node.Ancestors() {
				ancestorSpec := ancestor.Value()
				if ancestorSpec != nil {
					spec.PrependMiddlewares(ancestorSpec.middlewares...)
				}
			}
		}
	}
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pathkey := strings.TrimSuffix(req.URL.Path, "/")
	spec, ok := r.cachedPaths[pathkey]
	if !ok {
		spec, ok = r.paths.Get(pathkey)
		if !ok {
			http.NotFound(w, req)
			return
		}
		r.cachedPaths[pathkey] = spec
	}
	spec.ServeHTTP(w, req)
}
