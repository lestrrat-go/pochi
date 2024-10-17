package pochi

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/lestrrat-go/trie/v2"
)

// Router is an interface that represents a router: an object that can
// route HTTP requests to the appropriate handler based on the request path.
type Router interface {
	http.Handler
	MatchRoute(string) (*PathSpec, bool)
	Route(...*PathSpec) error
	Mount(string, Router) error
}

// TrieProvider is an interface that represents an object that can provide
// a trie structure. This is currently only used to call `Walk` on the router
type TrieProvider interface {
	Trie() *trie.Trie[string, string, *PathSpec]
}

type router struct {
	mu          sync.RWMutex
	paths       *pathtrie
	cachedPaths map[string]*PathSpec
}

func NewRouter() Router {
	return &router{
		paths:       newPathtrie(),
		cachedPaths: make(map[string]*PathSpec),
	}
}

func (r *router) Trie() *trie.Trie[string, string, *PathSpec] {
	return r.paths.impl
}

type RouteVisitor interface {
	Visit(string, *PathSpec) bool
}

type RouteVisitFunc func(string, *PathSpec) bool

func (f RouteVisitFunc) Visit(s string, spec *PathSpec) bool {
	return f(s, spec)
}

func recurseIterToPath(sb *strings.Builder, nodes []trie.Node[string, *PathSpec]) {
	if len(nodes) > 2 {
		recurseIterToPath(sb, nodes[1:])
		sb.WriteRune('/')
	}
	sb.WriteString(nodes[0].Key())
}

func iterToPath(nodes []trie.Node[string, *PathSpec]) string {
	var buf strings.Builder
	recurseIterToPath(&buf, nodes)
	return buf.String()
}

var errInvlidTrieProvider = errors.New("object does not provide a trie")

func ErrInvalidTrieProvider() error {
	return errInvlidTrieProvider
}

// Walk traverses the trie structure of the router, and calls the Visit method from
// the provided RouteVisitor for each path that has a handler attached to it.
// The path is the full path of the route, and the PathSpec is the associated PathSpec
// object.
//
// If the router does not implement a TrieProvider, this function will return an error
func Walk(r Router, v RouteVisitor) error {
	tp, ok := r.(TrieProvider)
	if !ok {
		return ErrInvalidTrieProvider()
	}
	trie.Walk(tp.Trie(), trie.VisitFunc[string, *PathSpec](func(n trie.Node[string, *PathSpec], _ trie.VisitMetadata) bool {
		if spec := n.Value(); spec != nil {
			v.Visit(iterToPath(n.Ancestors())+"/"+n.Key(), spec)
		}
		return true
	}))
	return nil
}

var errInvalidPath = errors.New("invalid path")

func ErrInvalidPath() error {
	return errInvalidPath
}

func (r *router) Route(specs ...*PathSpec) error {
	// perform checks _only_ first, so that we don't end up
	// with some path objects already processed and some erroring out
	for _, spec := range specs {
		if !strings.HasPrefix(spec.pattern, "/") {
			return fmt.Errorf("paths must be absolute %q: %w", spec.pattern, ErrInvalidPath())
		}

		// If for whatever reason this path object is already compiled,
		// we refuse to proceed
		if spec.compiled.Load() { // TODO: crete proper interface
			return fmt.Errorf("path %q is already compiled", spec.pattern)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return r.nlRoute(true, specs...)
}

func (r *router) nlRoute(recurse bool, specs ...*PathSpec) error {
	paths := make([]string, 0, len(specs))
	for _, spec := range specs {
		paths = append(paths, spec.pattern)
		r.paths.Put(spec.pattern, spec)
		node, ok := r.paths.GetNode(spec.pattern)
		if !ok {
			panic("failed to fetch node that we just inserted")
		}
		if spec.inheritMiddlewares {
			for _, ancestor := range node.Ancestors() {
				// Look for a child that has the "" key
				rootPath := ancestor.First()
				if rootPath.Key() != "" {
					continue
				}
				ancestorSpec := rootPath.Value()
				if ancestorSpec != nil {
					spec.InheritMiddlewares(ancestorSpec.directMiddlewares...)
				}
			}
		}
	}

	if !recurse {
		return nil
	}

	// It is possible to add /foo/ after /foo/bar, so we need to clear the cache, and
	// re-evaluate the paths
	var reroute []*PathSpec
	Walk(r, RouteVisitFunc(func(fullpath string, spec *PathSpec) bool {
		for _, p := range paths {
			if fullpath != p && strings.HasPrefix(fullpath, p) {
				delete(r.cachedPaths, fullpath)
				spec.invalidate()
				// r.paths.Delete(spec.pattern)
				reroute = append(reroute, spec)
				break
			}
		}
		return true
	}))

	if err := r.nlRoute(false, reroute...); err != nil {
		return fmt.Errorf(`failed to re-route after adding new paths: %w`, err)
	}

	return nil
}

func (r *router) MatchRoute(p string) (*PathSpec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nlMatchRoute(p)
}

func (r *router) nlMatchRoute(p string) (*PathSpec, bool) {
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
	r.mu.Lock()
	pathkey := req.URL.Path
	spec, ok := r.cachedPaths[pathkey]
	if !ok {
		spec, ok = r.nlMatchRoute(pathkey)
		if !ok || !spec.HasHandler() {
			r.mu.Unlock()
			http.NotFound(w, req)
			return
		}
		r.cachedPaths[pathkey] = spec
	}
	r.mu.Unlock()
	spec.ServeHTTP(w, req)
}

func (r *router) Mount(prefix string, router Router) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return Walk(router, RouteVisitFunc(func(fullpath string, spec *PathSpec) bool {
		p := MountPath(prefix, spec)
		if err := r.nlRoute(true, p); err != nil {
			return false
		}
		return true
	}))
}
