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
	require.Equal(t, "programming", feed.Title)
	require.Equal(t, "https://www.reddit.com/r/programming/", feed.Link)
	require.Len(t, feed.Entries, 25)

	first := feed.Entries[0]
	require.Equal(t, first.Title, "Dark Mode Coming to GitHub After 7 Years")
}

func TestYouTube(t *testing.T) {
	byt, err := ioutil.ReadFile("test-data/wandel.xml")
	require.Nil(t, err)

	feed, err := unmarshal(byt)
	require.Nil(t, err)
	require.Equal(t, "Matthias Wandel", feed.Title)
	require.Equal(t, "https://www.youtube.com/channel/UCckETVOT59aYw80B36aP9vw", feed.Link)
	require.Len(t, feed.Entries, 15)

	first := feed.Entries[0]
	require.Equal(t, "26\" bandsaw sawdust drawer and bottom enclosure", first.Title)
	require.Equal(t, "<div>Working on finishing up my 26\" bandsaw.  In this eposode, making the bottom enclosure and the sawdust drawer.  This directs nearly all the sawdust into the drawer, making for passive dust collection.\n\n\nhttp://woodgears.ca/big_bandsaw/bottom_enclosure.html</div><div><a href=\"https://www.youtube.com/v/9eRIUV94kgQ?version=3\"><img src=\"https://i2.ytimg.com/vi/9eRIUV94kgQ/hqdefault.jpg\" width=\"480\" height=\"360\" /></a></div>", string(first.Content))
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

func TestFileExists(t *testing.T) {
	exists := "readme.md"
	doesNotExist := "does-not-exist"
	require.True(t, fileExists(exists))
	require.False(t, fileExists(doesNotExist))
}
