# github.com/lestrrat-go/pochi ![](https://github.com/lestrrat-go/pochi/workflows/CI/badge.svg) [![Go Reference](https://pkg.go.dev/badge/github.com/lestrrat-go/pochi.svg)](https://pkg.go.dev/github.com/lestrrat-go/pochi)

WIP - Nothing in this repository is ready to be reviewed. Currently this is just a proof-of-concept, and a very rough one at that as well.

`pochi` is a CHI-inspired HTTP router. While it heavily borrows from the CHI
look and feel, the API tries to solve some of the problems that the authors
faced while using it in real-life applications.


# Basic usage

Pochi is comprised of "Path" objects (which, because of naming issues in Go, are called `PathSpec` objects).

You route requests to paths. As such, you will be interacting with `PathSpec` object most of the time.

```go
p := pochi.Path(`/foo`)
```

This creates a path spec that matches at a single request path `/foo`. `pochi` differentiates between
paths that end with a slash (`/`) and those that don't. Paths that end with a slash is treated as a
something that applies to everything under that path. For example, with a path ending in a slash,
all middlewares applied to that path is also applied to every path underneath it.

Paths that don't have a slash at the end is treated as a single endpoint, and has no such effect.

So assuming you have the following, where `p` is a path that ends with a slash, and `p2` is a path
that doesn't. Note that only `p` has a middleware enabled.

```go
p := pochi.Path(`/foo/`).
  Use(middleware.AccessLog())

p2 := pochi.Path(`/foo/bar`)
```

In this case `p2` will also have the accesslog middleware enabled.

One other effect of paths that ends with a slash is that it acts as a fallback handler for requests that don't match any other paths. But before that, we need to actually associate one or more handler with a path.

The following example cretes a path `/foo` with a handler that responds to 
requests that use HTTP GET method:

```go
p := pochi.Path(`/foo`).
    Get(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ...
    }))
```

In order to activate this path on a router, you need to register it.

```go
r := pochi.NewRouter()
r.Route(p)
```

And then you can register the router to a server expectig a `http.Handler`.

```go
srv := &http.Server{Handler: r}
```

# Problems with CHI that this module attempts to solve

## Easily dump all paths

Because of how CHI is structured, you can't just dump all registerd paths.
Now you can:

```go
r := pochi.NewRouter()
r.Routes(
    // initialization for paths below omitted
    pochi.Path("/foo"),
    pochi.Path("/bar"),
    pochi.Path("/baz"), 
)

r.Walk(pochi.RouteVisitFunc(func(path string, _ *pochi.PathSpec) {
    fmt.Println(path)
}
```

## Declare "nested" paths without resorting to nested closures

In CHI, you need to create a group of endpoints.

```go
// CHI
r.Router("/foo", func(r chi.Router) {
    r.Use(...)
    r.Get("/bar", ...)
    r.Get("/baz")
}
```

You could create another path that matches "/foo/..." from outside of this closure,
but it cannot benefit from the middlewares declared in the `r.Use(...)` statement.

`pochi` strictly works against paths, so this is no longer an issue. You can declare
handlers with arbitrary paths at any given point, and expect middlewares from
parent layers to be applied.

In the example below, both `/foo/bar` and `/foo/baz` inherit the middlewares
declared for `/foo`

```go
// pochi
r.Routes(
    pochi.Path("/foo").
        Use(...),
    pochi.Path("/foo/bar"),
    pochi.Path("/foo/baz"),
)
```

This means that you can have multiple stages of configuring the router.
For example, suppose a method does the above setting and return `r`.

Using CHI, you can't add new paths under `/foo` on that router expecting it to have the
same middlewares applied to it, but with this library

