package main

import (
	"io/ioutil"
	"testing"
	"time"

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

func TestParseDateNoTime(t *testing.T) {
	byt, err := ioutil.ReadFile("test-data/date-no-time.rss")
	require.Nil(t, err)

	feed, err := unmarshal(byt)
	require.Nil(t, err)
	require.Equal(t, len(feed.Entries), 1)

	updated := feed.Entries[0].Updated
	require.Equal(t, updated, time.Date(2020, 11, 23, 0, 0, 0, 0, time.UTC))
}
