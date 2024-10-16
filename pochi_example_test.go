package pochi_test

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/lestrrat-go/accesslog"
	"github.com/lestrrat-go/pochi"
	"github.com/lestrrat-go/pochi/middleware"
)

// This function creates an accesslog middleware that emits static values
// for the purposes of testing
func exampleAccessLog() *accesslog.Middleware {
	return middleware.AccessLog().
		// Use a static clock to get static output for testing
		Clock(accesslog.StaticClock(time.Time{})).
		Logger(
			slog.New(
				slog.NewJSONHandler(os.Stdout,
					&slog.HandlerOptions{
						ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
							switch a.Key {
							case slog.TimeKey:
								// replace time to get static output for testing
								return slog.Time(slog.TimeKey, time.Time{})
							case "remote_addr":
								// replace value to get static output for testing
								return slog.String("remote_addr", "127.0.0.1:99999")
							}
							return a
						},
					},
				),
			),
		)
}

func ExampleRouter() {
	r := pochi.NewRouter()

	if err := r.Route(
		pochi.Path("/foo/").
			Use(exampleAccessLog()),
		pochi.Path("/foo/regular").
			Get(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Hello, World! (inherits from /)")
			})),
		pochi.Path("/foo/nomiddlewares").
			InheritMiddlewares(false).
			Get(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Hello, World! (no accesslog)")
			})),
	); err != nil {
		fmt.Println(err)
		return
	}

	// Traverse all routes
	fmt.Println("--- All registered paths ---")
	pochi.Walk(r, pochi.RouteVisitFunc(func(fullpath string, spec *pochi.PathSpec) {
		fmt.Printf("Path: %s\n", fullpath)
	}))

	srv := httptest.NewServer(r)
	defer srv.Close()

	for _, path := range []string{"/foo/regular", "/foo/nomiddlewares"} {
		fmt.Printf("Issuing GET request to %q\n", path)
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

	// This would fail because the path "/foo/" does not have a handler
	res, err := srv.Client().Get(srv.URL + "/foo/bar")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		fmt.Printf("Expected status code 404, but got %d\n", res.StatusCode)
	}

	// OUTPUT:
	// --- All registered paths ---
	// Path: /foo/
	// Path: /foo/nomiddlewares
	// Path: /foo/regular
	// Issuing GET request to "/foo/regular"
	// {"time":"0001-01-01T00:00:00Z","level":"INFO","msg":"access","remote_addr":"127.0.0.1:99999","http_method":"GET","path":"/foo/regular","status":200,"body_bytes_sent":31,"http_referer":"","http_user_agent":"Go-http-client/1.1"}
	// 200
	// Hello, World! (inherits from /)
	// Issuing GET request to "/foo/nomiddlewares"
	// 200
	// Hello, World! (no accesslog)
}
