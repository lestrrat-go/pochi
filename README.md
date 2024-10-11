# pochi

WIP - Nothing in this repository is ready to be reviewed. Currently this is just a proof-of-concept, and a very rough one at that as well.

`pochi` is a CHI-inspired HTTP router. While it heavily borrows from the CHI
look and feel, the API tries to solve some of the problems that the authors
faced while using it in real-life applications.

# Easily dump all paths

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

# Declare "nested" paths without resorting to nested closures

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
        Use(...)
    pochi.Path("/foo/bar"),
    pochi.Path("/foo/baz"),
)
```