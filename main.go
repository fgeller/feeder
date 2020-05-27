package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// TODO timestamps
// TODO email template

type Feed struct {
	Title   string
	Updated time.Time
	Entries []*FeedEntry
}

type FeedEntry struct {
	Title   string
	Link    string
	ID      string
	Updated time.Time
	Content string
}

type RSSFeed struct { // v2
	XMLName       xml.Name  `xml:"rss"`
	Title         string    `xml:"channel>title"`
	Link          string    `xml:"channel>link"`
	LastBuildDate string    `xml:"channel>lastBuildDate"`
	Items         []RSSItem `xml:"channel>item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
}

func (f *RSSFeed) Feed() *Feed {
	ft, err := time.Parse(time.RFC1123Z, f.LastBuildDate)
	if err != nil {
		log.Fatalf("time parse str=%#v err=%v", f.LastBuildDate, err)
	}
	cf := &Feed{
		Title:   f.Title,
		Updated: ft,
		Entries: []*FeedEntry{},
	}
	for _, e := range f.Items {
		et, err := time.Parse(time.RFC1123Z, e.PubDate)
		if err != nil {
			log.Fatalf("time parse str=%#v err=%v", e.PubDate, err)
		}
		ce := &FeedEntry{
			Title:   e.Title,
			Link:    e.Link,
			ID:      e.GUID,
			Updated: et,
			Content: e.Description,
		}
		cf.Entries = append(cf.Entries, ce)
	}
	return cf
}

type AtomFeed struct {
	XMLName xml.Name     `xml:"feed"`
	Title   string       `xml:"title"`
	Link    AtomLink     `xml:"link"`
	Updated time.Time    `xml:"updated"`
	ID      string       `xml:"id"`
	Entries []*AtomEntry `xml:"entry"`
}

func (f *AtomFeed) String() string {
	return fmt.Sprintf("<%s updated=%s entries=%v>", f.Title, f.Updated, len(f.Entries))
}

func (f *AtomFeed) Feed() *Feed {
	cf := &Feed{
		Title:   f.Title,
		Updated: f.Updated,
		Entries: []*FeedEntry{},
	}
	for _, e := range f.Entries {
		cf.Entries = append(cf.Entries, &FeedEntry{
			Title:   e.Title,
			Link:    e.Link.HREF,
			ID:      e.ID,
			Updated: e.Updated,
			Content: e.Content,
		})
	}

	return cf
}

type AtomLink struct {
	HREF string `xml:"href,attr"`
}

type AtomEntry struct {
	Title   string    `xml:"title"`
	Link    AtomLink  `xml:"link"`
	Updated time.Time `xml:"updated"`
	ID      string    `xml:"id"`
	Content string    `xml:"content"`
}

func downloadFeed(url string) (*Feed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to request feed url=%s err=%w", url, err)
	}

	byt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body contents for feed url=%s err=%w", url, err)
	}
	defer resp.Body.Close()

	fmt.Printf("header:\n%s\n", byt[:512])

	isAtom := strings.Contains(string(byt[:46]), "<feed")

	if isAtom {
		var content AtomFeed
		err = xml.Unmarshal(byt, &content)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal atom content for feed url=%s err=%w", url, err)
		}
		return (&content).Feed(), nil
	} else {
		var content RSSFeed
		err = xml.Unmarshal(byt, &content)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rss content for feed url=%s err=%w", url, err)
		}
		return (&content).Feed(), nil
	}

	return nil, nil
}

func readFlags() (string, error) {
	var err error
	var cf string
	flags := flag.NewFlagSet("feeder", flag.ContinueOnError)
	flags.StringVar(&cf, "config", "", "Path to JSON config file (required).")

	err = flags.Parse(os.Args[1:])
	if err != nil {
		return "", err
	}

	if cf == "" {
		return "", fmt.Errorf("config is required.")
	}

	return cf, nil
}

type Config struct {
	TimestampFile string       `json:"timestamp-file"`
	Feeds         []ConfigFeed `json:"feeds"`
	Email         ConfigEmail  `json:"email"`
}

type ConfigEmail struct {
	From string     `json:"from"`
	Pass string     `json:"pass"`
	SMTP ConfigSMTP `json:"smtp"`
}

type ConfigSMTP struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type ConfigFeed struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

func readConfig() (*Config, error) {
	fn, err := readFlags()
	if err != nil {
		return nil, fmt.Errorf("failed to read flags: %w", err)
	}

	bt, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cf Config
	err = json.Unmarshal(bt, &cf)

	return &cf, err
}

func main() {
	var err error
	var cfg *Config

	cfg, err = readConfig()
	if err != nil {
		log.Fatal(err)
	}

	fs := []*Feed{}
	for _, fc := range cfg.Feeds {
		f, err := downloadFeed(fc.URL)
		if err != nil {
			log.Printf("dl err=%v", err)
		} else {
			fs = append(fs, f)
		}
	}
	log.Printf("downloaded %v feeds", len(fs))

	emailBody := makeEmailBody(fs)
	fmt.Printf("email body:\n%s\n", emailBody)
}

func makeEmailBody(feeds []*Feed) string {
	result := ""
	for _, f := range feeds {
		result += fmt.Sprintf("<h1>%s</h1>\n", f.Title)
		result += fmt.Sprintf("<h2>%s</h2>\n", f.Entries[0].Title)
		max := len(f.Entries[0].Content)
		if max > 128 {
			max = 128
		}
		result += f.Entries[0].Content[:max]
	}
	return result
}
