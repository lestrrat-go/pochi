package pochi_test

import (
	"fmt"
	"testing"

	"github.com/lestrrat-go/pochi"
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
}
