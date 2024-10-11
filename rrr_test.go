//go:build ignore

package rrr_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatterns(t *testing.T) {
	testcases := []struct {
		pattern string
		value   string
	}{
		{
			pattern: "/foo",
			value:   "/foo/bar",
		},
		{
			pattern: "/foo",
			value:   "/foo/bar/baz",
		},
	}
	for _, tc := range testcases {
		dir, file := filepath.Split(tc.value)
		strings.Split(tc.pattern, "/")

		match, err := filepath.Match(tc.pattern, tc.value)
		require.NoError(t, err, "should not error")
		require.True(t, match, "%q should match %q", tc.pattern, tc.value)
	}
}
