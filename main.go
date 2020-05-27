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

	"github.com/davecgh/go-spew/spew"
)

// TODO email template

type feed struct {
	name    string
	addr    string
	content AtomFeed
}

func newFeed(name, addr string) *feed {
	return &feed{name: name, addr: addr}
}

func (f *feed) download() error {
	resp, err := http.Get(f.addr)
	if err != nil {
		return fmt.Errorf("failed to request feed %v err=%w", f.name, err)
	}

	byt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body contents for feed %v err=%w", f.name, err)
	}
	defer resp.Body.Close()

	err = xml.Unmarshal(byt, &f.content)
	if err != nil {
		return fmt.Errorf("failed to unmarshal content for feed %v err=%w", f.name, err)
	}

	log.Printf("downloaded %v\n", f.content)
	return nil
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
	fmt.Println(err)
	spew.Dump(cfg)
}
