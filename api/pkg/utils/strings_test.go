package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceSymbols(t *testing.T) {
	s := "h!e.l/l>o$w/o\\r,l<d"

	assert.Equal(t, "helloworld", ReplaceSymbols(s), "All symbols are replaced")
}

func TestReplaceSpace(t *testing.T) {
	s := "hello world"

	assert.Equal(t, "helloworld", ReplaceSymbols(s), "A single space is replaced")
}

func TestReplaceMultipleSpaces(t *testing.T) {
	s := "he  l  l  o w               o r l        d"

	assert.Equal(t, "helloworld", ReplaceSymbols(s), "Multiple spaces in sequence are replaced")
}
