package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTakeOnRules(t *testing.T) {
	byt, err := ioutil.ReadFile("test-data/take-on-rules.atom")
	require.Nil(t, err)

	_, err = unmarshal(byt)
	require.Nil(t, err)
}

func TestNotUtf8(t *testing.T) {
	byt, err := ioutil.ReadFile("test-data/not-utf8.rss")
	require.Nil(t, err)

	_, err = unmarshal(byt)
	require.Nil(t, err)
}
