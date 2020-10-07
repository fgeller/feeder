package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
	"gopkg.in/yaml.v2"
)

// Feed represents a downloaded news feed
type Feed struct {
	Title   string
	ID      string
	Link    string
	Updated time.Time
	Entries []*FeedEntry
}

// FeedEntry represents a a downloaded news feed entry
type FeedEntry struct {
	Title   string
	Link    string
	ID      string
	Updated time.Time
	Content template.HTML
}

func (e *FeedEntry) Copy() *FeedEntry {
	return &FeedEntry{
		Title:   e.Title,
		Link:    e.Link,
		ID:      e.ID,
		Updated: e.Updated,
		Content: e.Content,
	}
}

type RSSFeed struct { // v2
	XMLName       xml.Name  `xml:"rss"`
	Title         string    `xml:"channel>title"`
	Links         []Link    `xml:"channel>link"`
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
	if len(f.Links) == 0 {
		log.Fatalf("missing link on feed %#v", f.Title)
	}

	cf := &Feed{
		ID:      f.Links[0].HRef, // 🤨
		Title:   f.Title,
		Link:    f.Links[0].HRef,
		Entries: []*FeedEntry{},
	}

	var err error
	if f.LastBuildDate != "" {
		cf.Updated, err = time.Parse(time.RFC1123Z, f.LastBuildDate)
		if err != nil {
			log.Fatalf("time parse feed title=%v str=%#v err=%v", f.Title, f.LastBuildDate, err)
		}
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
			Content: template.HTML(e.Description),
		}
		cf.Entries = append(cf.Entries, ce)
	}
	return cf
}

type AtomFeed struct {
	XMLName xml.Name     `xml:"feed"`
	Title   string       `xml:"title"`
	Link    Link         `xml:"link"`
	Updated time.Time    `xml:"updated"`
	ID      string       `xml:"id"`
	Entries []*AtomEntry `xml:"entry"`
}

func (f *AtomFeed) Feed() *Feed {
	cf := &Feed{
		ID:      f.ID,
		Title:   f.Title,
		Link:    f.Link.HRef,
		Updated: f.Updated,
		Entries: []*FeedEntry{},
	}
	for _, e := range f.Entries {
		cf.Entries = append(cf.Entries, &FeedEntry{
			Title:   e.Title,
			Link:    e.Link.HRef,
			ID:      e.ID,
			Updated: e.Updated,
			Content: template.HTML(e.Content),
		})
	}

	return cf
}

// Link enables us to unmarshal Atom and plain link tags
type Link struct {
	XMLName xml.Name `xml:"link"`
	HRef    string
}

func (l *Link) UnmarshalXML(d *xml.Decoder, el xml.StartElement) error {
	var s string
	err := d.DecodeElement(&s, &el)
	if err != nil {
		return err
	}

	_, err = url.ParseRequestURI(s)
	if err == nil {
		l.HRef = s
		return nil
	}

	if len(el.Attr) > 0 {
		for _, a := range el.Attr {
			if a.Name.Local == "href" {
				_, err = url.ParseRequestURI(a.Value)
				if err == nil {
					l.HRef = a.Value
					return nil
				}
			}
		}
	}

	return fmt.Errorf("found no href content in link element %#v", el)
}

type AtomEntry struct {
	Title   string    `xml:"title"`
	Link    Link      `xml:"link"`
	Updated time.Time `xml:"updated"`
	ID      string    `xml:"id"`
	Content string    `xml:"content"`
}

func downloadFeed(url string) (*Feed, error) {
	log.Printf("downloading feed %#v\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to request feed url=%s err=%w", url, err)
	}

	byt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body contents for feed url=%s err=%w", url, err)
	}
	defer resp.Body.Close()

	isAtom := strings.Contains(string(byt[:128]), "<feed")

	if isAtom {
		var content AtomFeed
		err = xml.Unmarshal(byt, &content)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal atom content for feed url=%s err=%w", url, err)
		}
		return (&content).Feed(), nil
	}

	var content RSSFeed
	err = xml.Unmarshal(byt, &content)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal rss content for feed url=%s err=%w", url, err)
	}
	return (&content).Feed(), nil
}

func readFlags() (string, error) {
	var err error
	var cf string
	flags := flag.NewFlagSet("feeder", flag.ContinueOnError)
	flags.StringVar(&cf, "config", "", "Path to config file (required).")

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
	TimestampFile     string       `yaml:"timestamp-file"`
	EmailTemplateFile string       `yaml:"email-template-file"`
	Feeds             []ConfigFeed `yaml:"feeds"`
	Email             ConfigEmail  `yaml:"email"`
}

type ConfigEmail struct {
	From string     `yaml:"from"`
	SMTP ConfigSMTP `yaml:"smtp"`
}

type ConfigSMTP struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type ConfigFeed struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Disabled bool   `yaml:"disabled"`
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
	err = yaml.Unmarshal(bt, &cf)

	return &cf, err
}

func failOnErr(err error) {
	if err != nil {
		cfg, nerr := readConfig()
		if nerr == nil {
			cf := cfg.Email
			m := gomail.NewMessage()
			m.SetHeader("From", cf.From)
			m.SetHeader("To", cf.From)
			m.SetHeader("Subject", "feeder failure")
			m.SetBody("text/plain", err.Error())

			d := gomail.NewDialer(cf.SMTP.Host, cf.SMTP.Port, cf.SMTP.User, cf.SMTP.Pass)
			log.Printf("tried to send failure email err=%v", d.DialAndSend(m))
		}
		log.Fatal(err)
	}
}

func sendEmail(cfg ConfigEmail, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", cfg.From)
	m.SetHeader("Subject", fmt.Sprintf("feeder update: %s", time.Now().Format("2006-01-02 15:04")))
	m.SetBody("text/html", body)

	d := gomail.NewDialer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.User, cfg.SMTP.Pass)
	return d.DialAndSend(m)
}

func downloadFeeds(cs []ConfigFeed) ([]*Feed, error) {
	fs := []*Feed{}
	for _, fc := range cs {
		if fc.Disabled {
			continue
		}
		f, err := downloadFeed(fc.URL)
		if err != nil {
			return fs, fmt.Errorf("failed to download feed err=%w", err)
		} else {
			fs = append(fs, f)
		}
	}
	return fs, nil
}

func pickNewData(fs []*Feed, ts map[string]time.Time) []*Feed {
	limitPerFeed := 3
	result := []*Feed{}
	for _, f := range fs {
		nf := &Feed{Title: f.Title, ID: f.ID, Link: f.Link, Updated: f.Updated, Entries: []*FeedEntry{}}
		lt, seen := ts[f.ID]
		for _, e := range f.Entries {
			if !seen || e.Updated.After(lt) {
				nf.Entries = append(nf.Entries, e.Copy())
				if len(nf.Entries) >= limitPerFeed {
					break
				}
			}
		}
		if len(nf.Entries) > 0 {
			result = append(result, nf)
		}
	}
	return result
}

func updateTimestamps(ts map[string]time.Time, nd []*Feed) {
	for _, f := range nd {
		_, ok := ts[f.ID]
		if !ok {
			ts[f.ID] = f.Entries[0].Updated
		}
		for _, e := range f.Entries {
			if e.Updated.After(ts[f.ID]) {
				ts[f.ID] = e.Updated
			}
		}
	}
}

func readTimestamps(fn string) (map[string]time.Time, error) {
	var err error
	var result map[string]time.Time
	var bt []byte
	var fh *os.File

	fh, err = os.OpenFile(fn, os.O_CREATE, 0677)
	if err != nil {
		return nil, fmt.Errorf("failed to open timestamps file %#v err=%w", fn, err)
	}

	bt, err = ioutil.ReadAll(fh)
	if err != nil {
		return nil, fmt.Errorf("failed to read timestamps file %#v err=%w", fn, err)
	}

	if len(bt) == 0 {
		return map[string]time.Time{}, nil
	}

	err = yaml.Unmarshal(bt, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal timestamps %#v file err=%w", fn, err)
	}

	return result, nil
}

func writeTimestamps(fn string, ts map[string]time.Time) error {
	var err error
	var bt []byte

	bt, err = yaml.Marshal(ts)
	if err != nil {
		return fmt.Errorf("failed to marshal timestamps err=%w", err)
	}

	err = ioutil.WriteFile(fn, bt, 0677)
	if err != nil {
		return fmt.Errorf("failed to write timestamps file err=%w", err)
	}

	return nil
}

// FormatTime prints a time with layout "2006-01-02 15:04 MST"
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04 MST")
}

var defaultEmailTemplate = `
{{ range .}}
<h1 style="background:#f5f5f5;padding:0.5rem;border-radius:3px;"><a href="{{ .Link }}" style="text-decoration:none;">{{ .Title }}</a></h1>
{{ range .Entries }}
<h2 style="background:#f5f5f5;padding:0.5rem;border-radius:3px;"><a href="{{ .Link }}" style="text-decoration:none;">{{ .Title }}</a><span style="font-size:0.75rem;margin-left:1rem;">{{ FormatTime .Updated }}</span></h2>
<div>
  {{ .Content }}
</div>
{{ end }}
{{ end }}
`

func readEmailTemplate(fn string) (string, error) {
	if fn == "" {
		return defaultEmailTemplate, nil
	}

	bt, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", fmt.Errorf("failed to read email template file %#v err=%w", fn, err)
	}

	return string(bt), nil
}

func makeEmailBody(feeds []*Feed, emailTemplate string) (string, error) {
	fs := template.FuncMap{"FormatTime": FormatTime}
	tmpl, err := template.New("email").Funcs(fs).Parse(emailTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template err=%w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, feeds)
	if err != nil {
		return "", fmt.Errorf("failed to execute template err=%w", err)
	}

	return buf.String(), nil
}

func countEntries(fs []*Feed) int {
	c := 0
	for _, f := range fs {
		c += len(f.Entries)
	}
	return c
}

func main() {
	var err error
	var cfg *Config
	var ts map[string]time.Time
	var fs, nd []*Feed
	var et string

	cfg, err = readConfig()
	failOnErr(err)
	log.Printf("read config\n")

	ts, err = readTimestamps(cfg.TimestampFile)
	failOnErr(err)
	log.Printf("read timestamps from %#v\n", cfg.TimestampFile)

	et, err = readEmailTemplate(cfg.EmailTemplateFile)
	failOnErr(err)

	fs, err = downloadFeeds(cfg.Feeds)
	failOnErr(err)
	log.Printf("donwloaded %v feeds\n", len(fs))

	nd = pickNewData(fs, ts)
	if len(nd) == 0 {
		log.Printf("found no new entries")
		return
	}
	log.Printf("found %v new entries\n", countEntries(nd))

	emailBody, err := makeEmailBody(nd, et)
	failOnErr(err)

	err = sendEmail(cfg.Email, emailBody)
	failOnErr(err)
	log.Printf("sent email\n")

	updateTimestamps(ts, nd)
	err = writeTimestamps(cfg.TimestampFile, ts)
	failOnErr(err)
	log.Printf("wrote updated timestamps to %#v\n", cfg.TimestampFile)
}
