package pochi

import (
	"strings"

	"github.com/lestrrat-go/trie/v2"
)

type impl = trie.Trie[string, string, *PathSpec]

type pathtrie struct {
	*impl
}

func newPathtrie() *pathtrie {
	return &pathtrie{
		impl: trie.New[string, string, *PathSpec](pathTokenizer{}),
	}
}

type pathTokenizer struct{}

func (pathTokenizer) Tokenize(s string) ([]string, error) {
	return strings.Split(s, "/"), nil
}
