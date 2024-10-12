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

func ExampleRouter() {
	r := pochi.NewRouter()

	r.Route(
		pochi.Path("/").
			Use(middleware.AccessLog().
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
				),
			).
			Get(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Hello, World!")
			})),
		pochi.Path("/regular").
			Get(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Hello, World! (inherits from /)")
			})),
		pochi.Path("/noaccesslog").
			InheritMiddlewares(false).
			Get(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Hello, World! (no accesslog)")
			})),
	)

	// Traverse all routes
	fmt.Println("--- All registered paths ---")
	r.Walk(pochi.RouteVisitFunc(func(fullpath string, spec *pochi.PathSpec) {
		fmt.Printf("Path: %s\n", fullpath)
	}))

	srv := httptest.NewServer(r)
	defer srv.Close()

	for _, path := range []string{"/", "/regular", "/noaccesslog", "/test/"} {
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

	// OUTPUT:
	// --- All registered paths ---
	// Path: /
	// Path: /noaccesslog
	// Path: /regular
	// Issuing GET request to "/"
	// {"time":"0001-01-01T00:00:00Z","level":"INFO","msg":"access","remote_addr":"127.0.0.1:99999","http_method":"GET","path":"/","status":200,"body_bytes_sent":13,"http_referer":"","http_user_agent":"Go-http-client/1.1"}
	// 200
	// Hello, World!
	// Issuing GET request to "/regular"
	// 200
	// Hello, World! (inherits from /)
	// Issuing GET request to "/noaccesslog"
	// 200
	// Hello, World! (no accesslog)
	// Issuing GET request to "/test/"
	// {"time":"0001-01-01T00:00:00Z","level":"INFO","msg":"access","remote_addr":"127.0.0.1:99999","http_method":"GET","path":"/test/","status":200,"body_bytes_sent":13,"http_referer":"","http_user_agent":"Go-http-client/1.1"}
	// 200
	// Hello, World!
}
