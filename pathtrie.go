package pochi

import (
	"iter"
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

func (pathTokenizer) Tokenize(s string) (iter.Seq[string], error) {
	comps := strings.Split(s, "/")
	return func(yield func(string) bool) {
		for _, c := range comps {
			if !yield(c) {
				break
			}
		}
	}, nil
}
