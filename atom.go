package main

import (
	"encoding/xml"
	"fmt"
	"time"
)

type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Link    AtomLink    `xml:"link"`
	Updated time.Time   `xml:"updated"`
	ID      string      `xml:"id"`
	Entries []AtomEntry `xml:"entry"`
}

func (f AtomFeed) String() string {
	return fmt.Sprintf("<%s updated=%s entries=%v>", f.Title, f.Updated, len(f.Entries))
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
