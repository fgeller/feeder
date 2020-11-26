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

func TestReddit(t *testing.T) {
	byt, err := ioutil.ReadFile("test-data/rprogramming.atom")
	require.Nil(t, err)

	feed, err := unmarshal(byt)
	require.Nil(t, err)
	require.Equal(t, feed.Title, "programming")
	require.Equal(t, feed.Link, "https://www.reddit.com/r/programming/")
	require.Equal(t, len(feed.Entries), 25)

	first := feed.Entries[0]
	require.Equal(t, first.Title, "Dark Mode Coming to GitHub After 7 Years")
}

func TestNotUtf8(t *testing.T) {
	byt, err := ioutil.ReadFile("test-data/not-utf8.rss")
	require.Nil(t, err)

	_, err = unmarshal(byt)
	require.Nil(t, err)
}
