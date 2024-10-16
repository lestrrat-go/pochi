package pochi_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lestrrat-go/pochi"
	"github.com/lestrrat-go/pochi/middleware"
	"github.com/stretchr/testify/require"
)

func Test_sanity(t *testing.T) {
	t.Run("Inherited path", func(t *testing.T) {
		r := pochi.NewRouter()
		require.NoError(t, r.Route(pochi.Path("/")), "route should succeed")

		// This should work
		for _, path := range []string{"/", "/foo"} {
			t.Run(fmt.Sprintf("path = %q", path), func(t *testing.T) {
				ps, ok := r.MatchRoute(path)
				require.True(t, ok, "path should exist")
				require.NotNil(t, ps, "path should not be nil")
			})
		}
	})
	t.Run("Non-inherited path", func(t *testing.T) {
		r := pochi.NewRouter()
		require.NoError(t, r.Route(pochi.Path("/foo")), "route should succeed")

		testcases := []struct {
			Path  string
			Error bool
		}{
			{Path: "/foo", Error: false},
			{Path: "/foo/bar", Error: true},
		}

		for _, tc := range testcases {
			t.Run(fmt.Sprintf("path = %q", tc.Path), func(t *testing.T) {
				ps, ok := r.MatchRoute(tc.Path)
				if tc.Error {
					require.False(t, ok, "path should not exist")
				} else {
					require.True(t, ok, "path should exist")
					require.NotNil(t, ps, "path should not be nil")
				}
			})
		}
	})
	t.Run("relative paths", func(t *testing.T) {
		r := pochi.NewRouter()
		require.Error(t, r.Route(pochi.Path("foo")), "relative paths should not be allowed")
	})

	t.Run("middleware order", func(t *testing.T) {
		addcount := func(i int) pochi.Middleware {
			return pochi.MiddlewareFunc(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
					fmt.Fprintf(w, "%d\n", i)
				})
			})
		}
		r := pochi.NewRouter()
		require.NoError(t, r.Route(
			pochi.Path("/foo/").
				Use(
					addcount(1),
					addcount(2),
					addcount(3),
				),
			pochi.Path("/foo/bar").
				Use(
					addcount(4),
					addcount(5),
					addcount(6),
				).Get(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "7\n")
			})),
		), "route should succeed")

		srv := httptest.NewServer(r)
		defer srv.Close()

		res, err := srv.Client().Get(srv.URL + "/foo/bar")
		require.NoError(t, err, "GET should succeed")
		defer res.Body.Close()

		require.Equal(t, http.StatusOK, res.StatusCode, "status code should be 200")
		buf, err := io.ReadAll(res.Body)
		require.NoError(t, err, "reading response body should succeed")
		require.Equal(t, "7\n6\n5\n4\n3\n2\n1\n", string(buf), "response body should match")

	})

	t.Run("Mount", func(t *testing.T) {
		r1 := pochi.NewRouter()
		r1.Route(
			pochi.Path("/foo").
				Use(middleware.AccessLog()),
		)

		printpath := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", r.URL.Path)
		})
		r2 := pochi.NewRouter()
		r2.Route(
			pochi.Path("/bar").
				Get(printpath),
			pochi.Path("/baz").
				Get(printpath),
		)

		require.NoError(t, r1.Mount(`/foo`, r2), "mount should succeed")

		srv := httptest.NewServer(r1)
		defer srv.Close()

		for _, path := range []string{"/foo/bar", "/foo/baz"} {
			t.Run(fmt.Sprintf("path = %q", path), func(t *testing.T) {
				res, err := srv.Client().Get(srv.URL + path)
				require.NoError(t, err, "GET should succeed")
				defer res.Body.Close()

				require.Equal(t, http.StatusOK, res.StatusCode, "status code should be 200")
				buf, err := io.ReadAll(res.Body)
				require.NoError(t, err, "reading response body should succeed")
				require.Equal(t, path, string(buf), "response body should match")
			})
		}
	})
}
