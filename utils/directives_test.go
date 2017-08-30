package utils

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirectivesRegistry(t *testing.T) {
	AddDirective(func(in io.Reader) []byte {
		b, err := ioutil.ReadAll(in)
		require.Nil(t, err)
		r := string(b) + "test1"
		return []byte(r)
	})
	AddDirective(func(in io.Reader) []byte {
		b, err := ioutil.ReadAll(in)
		require.Nil(t, err)
		r := string(b) + "test2"
		return []byte(r)
	})
	input := []byte("start:")
	result := ApplyDirectives(input)
	require.Equal(t, "start:test1test2", result)

	input2 := []byte("start:\n")
	result2 := ApplyDirectives(input2)
	require.Equal(t, "start:\ntest1test2", result2)
}
