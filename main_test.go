package main

import (
	"io/ioutil"
	"net/url"
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

func TestSubstituteRelativeImageSrc(t *testing.T) {
	orig := `src="/plus/misc/images/her-soundtrack.jpg"`
	expected := `src="http://kottke.org/plus/misc/images/her-soundtrack.jpg"`

	byt, err := ioutil.ReadFile("test-data/kottke-entry.html")
	require.Nil(t, err)
	require.Contains(t, string(byt), orig, "entry should contain relative url")

	bu, err := url.Parse("http://kottke.org/")
	res, err := absolutifyHTML(string(byt), bu)
	require.Contains(t, string(res), expected, "relative url should be replaced by absolute one")
}

func TestSubstituteRelativeAHref(t *testing.T) {
	orig := `href="/bradfitz"`
	expected := `href="https://github.com/bradfitz"`

	byt, err := ioutil.ReadFile("test-data/github-entry.html")
	require.Nil(t, err)
	require.Contains(t, string(byt), orig, "entry should contain relative url")
	require.NotContains(t, string(byt), expected, "absolute url should not yet be present")

	bu, err := url.Parse("https://github.com/")
	res, err := absolutifyHTML(string(byt), bu)
	require.Contains(t, string(res), expected, "relative url should be replaced by absolute one")
	require.NotContains(t, string(res), orig, "relative url should not be present anymore")
}
