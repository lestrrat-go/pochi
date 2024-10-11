package pochi_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/lestrrat-go/pochi"
	"github.com/lestrrat-go/pochi/middleware"
)

func ExampleRouter() {
	r := pochi.NewRouter()

	r.Route(
		pochi.Path("/").
			Use(middleware.AccessLog()).
			Get(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Hello, World!")
			})),
		pochi.Path("/noaccesslog").
			InheritMiddlewares(false).
			Get(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Hello, World! (no accesslog)")
			})),
	)

	// Traverse all routes
	r.Walk(pochi.RouteVisitFunc(func(fullpath string, spec *pochi.PathSpec) {
		fmt.Printf("Path: %s\n", fullpath)
	}))

	srv := httptest.NewServer(r)
	defer srv.Close()

	for _, path := range []string{"/", "/noaccesslog"} {
		res, err := srv.Client().Get(srv.URL + path)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer res.Body.Close()

		fmt.Println(res.StatusCode)
		buf, _ := io.ReadAll(res.Body)
		fmt.Println(string(buf))
	}

	// OUTPUT:
	// Path: /
	// Path: /noaccesslog
	// HTTP GET /
	// 200
	// Hello, World!
	// 200
	// Hello, World! (no accesslog)
}
