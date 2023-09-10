package main

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUnmarshal_RDF(t *testing.T) {
	byt, err := os.ReadFile("test-data/slashdotMain.xml")
	require.Nil(t, err)

	f, err := unmarshal(byt)
	require.Nil(t, err)

	require.Equal(t, "Slashdot", f.Title)
	require.Equal(t, "https://slashdot.org/", f.Link)
	require.Len(t, f.Entries, 15)
	require.Equal(t, time.Date(2022, 7, 28, 10, 52, 17, 0, time.UTC).Unix(), f.Updated.Unix())

	fst := f.Entries[0]
	require.Equal(t, "Charter Told To Pay $7.3 Billion In Damages After Cable Installer Murders Grandmother", fst.Title)
	require.Equal(t, "https://yro.slashdot.org/story/22/07/27/2124200/charter-told-to-pay-73-billion-in-damages-after-cable-installer-murders-grandmother?utm_source=rss1.0mainlinkanon&utm_medium=feed", fst.Link)
	require.Equal(t, "https://yro.slashdot.org/story/22/07/27/2124200/charter-told-to-pay-73-billion-in-damages-after-cable-installer-murders-grandmother?utm_source=rss1.0mainlinkanon&utm_medium=feed", fst.ID)
	require.Equal(t, time.Date(2022, 7, 28, 10, 0, 0, 0, time.UTC).Unix(), fst.Updated.Unix())
}

func TestTakeOnRules(t *testing.T) {
	byt, err := os.ReadFile("test-data/take-on-rules.atom")
	require.Nil(t, err)

	f, err := unmarshal(byt)
	require.Nil(t, err)

	require.Equal(t, "https://takeonrules.com/", f.Link)
}

func TestReddit(t *testing.T) {
	byt, err := os.ReadFile("test-data/rprogramming.atom")
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
	byt, err := os.ReadFile("test-data/wandel.xml")
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
	byt, err := os.ReadFile("test-data/not-utf8.rss")
	require.Nil(t, err)

	_, err = unmarshal(byt)
	require.Nil(t, err)
}

func TestParseDateNoTime(t *testing.T) {
	byt, err := os.ReadFile("test-data/date-no-time.rss")
	require.Nil(t, err)

	feed, err := unmarshal(byt)
	require.Nil(t, err)
	require.Equal(t, len(feed.Entries), 1)

	updated := feed.Entries[0].Updated
	require.Equal(t, updated, time.Date(2020, 11, 23, 0, 0, 0, 0, time.UTC))
}

func TestParseTime(t *testing.T) {
	data := []struct {
		raw      string
		expected time.Time
		err      error
	}{
		{
			raw:      "Mon, 2 March 2020 12:00:00 CET",
			expected: time.Date(2020, 3, 2, 12, 0, 0, 0, time.UTC),
			err:      nil,
		},
		{
			raw:      "Tue, 2 Mar 2021 06:50:00 +1300",
			expected: time.Date(2021, 3, 1, 17, 50, 0, 0, time.UTC),
			err:      nil,
		},
	}

	for _, d := range data {
		actual, err := parseTime(d.raw)
		require.Equal(t, d.err, err)
		require.Equal(t, int64(0), d.expected.Unix()-actual.Unix())
	}
}

func TestSubstituteRelativeImageSrc(t *testing.T) {
	orig := `src="/plus/misc/images/her-soundtrack.jpg"`
	expected := `src="http://kottke.org/plus/misc/images/her-soundtrack.jpg"`

	byt, err := os.ReadFile("test-data/kottke-entry.html")
	require.Nil(t, err)
	require.Contains(t, string(byt), orig, "entry should contain relative url")

	bu, err := url.Parse("http://kottke.org/")
	res, err := absolutifyHTML(string(byt), bu)
	require.Contains(t, string(res), expected, "relative url should be replaced by absolute one")
}

func TestSubstituteRelativeAHref(t *testing.T) {
	orig := `href="/bradfitz"`
	expected := `href="https://github.com/bradfitz"`

	byt, err := os.ReadFile("test-data/github-entry.html")
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

func TestPickNewData(t *testing.T) {
	td := map[string]struct {
		feeds        []*Feed
		limitPerFeed int
		timestamps   map[string]time.Time
		expected     []*Feed
	}{
		"one new entry": {
			feeds: []*Feed{
				{
					Title:   "Test Feed",
					ID:      "5db01937",
					Link:    "http://example.com",
					Updated: time.Date(2022, 7, 23, 1, 2, 3, 0, time.UTC),
					Entries: []*FeedEntry{
						{
							Title:   "Old Entry",
							Link:    "http://example.com/old",
							ID:      "5db01937-old",
							Updated: time.Date(2022, 7, 22, 1, 2, 3, 0, time.UTC),
						},
						{
							Title:   "Latest Hidden Entry",
							Link:    "http://example.com/new",
							ID:      "5db01937-new",
							Updated: time.Date(2022, 7, 23, 1, 2, 3, 0, time.UTC),
						},
					},
				},
			},
			limitPerFeed: 1,
			timestamps: map[string]time.Time{
				"5db01937": time.Date(2022, 7, 22, 1, 2, 3, 0, time.UTC),
			},
			expected: []*Feed{
				{
					Title:   "Test Feed",
					ID:      "5db01937",
					Link:    "http://example.com",
					Updated: time.Date(2022, 7, 23, 1, 2, 3, 0, time.UTC),
					Entries: []*FeedEntry{
						{
							Title:   "Latest Hidden Entry",
							Link:    "http://example.com/new",
							ID:      "5db01937-new",
							Updated: time.Date(2022, 7, 23, 1, 2, 3, 0, time.UTC),
						},
					},
				},
			},
		},
		"initial pick is latest entry": {
			feeds: []*Feed{
				{
					Title:   "Test Feed",
					ID:      "5db01937",
					Link:    "http://example.com",
					Updated: time.Date(2022, 7, 23, 1, 2, 3, 0, time.UTC),
					Entries: []*FeedEntry{
						{
							Title:   "Old Entry",
							Link:    "http://example.com/old",
							ID:      "5db01937-old",
							Updated: time.Date(2022, 7, 22, 1, 2, 3, 0, time.UTC),
						},
						{
							Title:   "Latest Hidden Entry",
							Link:    "http://example.com/new",
							ID:      "5db01937-new",
							Updated: time.Date(2022, 7, 23, 1, 2, 3, 0, time.UTC),
						},
					},
				},
			},
			limitPerFeed: 1,
			timestamps:   map[string]time.Time{},
			expected: []*Feed{
				{
					Title:   "Test Feed",
					ID:      "5db01937",
					Link:    "http://example.com",
					Updated: time.Date(2022, 7, 23, 1, 2, 3, 0, time.UTC),
					Entries: []*FeedEntry{
						{
							Title:   "Latest Hidden Entry",
							Link:    "http://example.com/new",
							ID:      "5db01937-new",
							Updated: time.Date(2022, 7, 23, 1, 2, 3, 0, time.UTC),
						},
					},
				},
			},
		},
		"oldest entry first": {
			feeds: []*Feed{
				{
					Title:   "Test Feed",
					ID:      "5db01937",
					Link:    "http://example.com",
					Updated: time.Date(2022, 7, 23, 3, 2, 3, 0, time.UTC),
					Entries: []*FeedEntry{
						{
							Title:   "Entry 3",
							Link:    "http://example.com/3",
							ID:      "5db01937-3",
							Updated: time.Date(2022, 7, 22, 3, 2, 3, 0, time.UTC),
						},
						{
							Title:   "Entry 1",
							Link:    "http://example.com/1",
							ID:      "5db01937-1",
							Updated: time.Date(2022, 7, 22, 1, 2, 3, 0, time.UTC),
						},
						{
							Title:   "Entry 2",
							Link:    "http://example.com/2",
							ID:      "5db01937-2",
							Updated: time.Date(2022, 7, 22, 2, 2, 3, 0, time.UTC),
						},
					},
				},
			},
			limitPerFeed: 3,
			timestamps:   map[string]time.Time{},
			expected: []*Feed{
				{
					Title:   "Test Feed",
					ID:      "5db01937",
					Link:    "http://example.com",
					Updated: time.Date(2022, 7, 23, 3, 2, 3, 0, time.UTC),
					Entries: []*FeedEntry{
						{
							Title:   "Entry 1",
							Link:    "http://example.com/1",
							ID:      "5db01937-1",
							Updated: time.Date(2022, 7, 22, 1, 2, 3, 0, time.UTC),
						},
						{
							Title:   "Entry 2",
							Link:    "http://example.com/2",
							ID:      "5db01937-2",
							Updated: time.Date(2022, 7, 22, 2, 2, 3, 0, time.UTC),
						},
						{
							Title:   "Entry 3",
							Link:    "http://example.com/3",
							ID:      "5db01937-3",
							Updated: time.Date(2022, 7, 22, 3, 2, 3, 0, time.UTC),
						},
					},
				},
			},
		},
	}

	for tn, tc := range td {
		actual := pickNewData(tc.feeds, tc.limitPerFeed, tc.timestamps)
		require.Equal(t, tc.expected, actual, tn)
	}
}
