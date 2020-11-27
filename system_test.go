package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

type cmd struct {
	in string
}

func newCmd() *cmd                  { return &cmd{} }
func (c *cmd) stdIn(in string) *cmd { c.in = in; return c }
func (c *cmd) run(name string, args ...string) (int, string, string) {
	cmd := exec.Command(name, args...)

	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	if len(c.in) > 0 {
		cmd.Stdin = strings.NewReader(c.in)
	}

	_ = cmd.Run()
	status := cmd.ProcessState.Sys().(syscall.WaitStatus)

	strOut := stdOut.String()
	strErr := stdErr.String()

	return status.ExitStatus(), strOut, strErr
}

func build(t *testing.T) {
	var status int

	status, _, _ = newCmd().run("make", "build")
	require.Zero(t, status)

	status, _, _ = newCmd().run("ls", "feeder")
	require.Zero(t, status)
}

func TestSystem(t *testing.T) {
	build(t)

	feedsFile := "./test-data/feeds.yml"
	feeds := `- name: The Go Blog
  url: https://blog.golang.org/feed.atom
  disabled: false
`
	err := ioutil.WriteFile(feedsFile, []byte(feeds), 0677)
	require.Nil(t, err)

	var status int
	var stdOut, stdErr string

	//
	// feeder -config test-data/subscribe-cfg.yml -subscribe https://blog.golang.org/
	//
	status, stdOut, stdErr = newCmd().
		run("./feeder",
			"-config", "test-data/subscribe-cfg.yml",
			"-subscribe", "https://blog.golang.org/",
		)
	fmt.Printf(">> feeder -config test-data/subscribe-cfg.yml -subscribe https://blog.golang.org/ stdout:\n%s\n", stdOut)
	fmt.Printf(">> feeder -config test-data/subscribe-cfg.yml -subscribe https://blog.golang.org/ stderr:\n%s\n", stdErr)
	require.Zero(t, status)

	fs, err := readFeedsConfig("test-data/feeds.yml")
	require.Nil(t, err)
	require.Len(t, fs, 1)
	expected0 := &ConfigFeed{Name: "The Go Blog", URL: "https://blog.golang.org/feed.atom"}
	require.Equal(t, expected0, fs[0])

	//
	// feeder -config test-data/subscribe-cfg.yml -subscribe https://www.kottke.org/
	//
	status, stdOut, stdErr = newCmd().
		run("./feeder",
			"-config", "test-data/subscribe-cfg.yml",
			"-subscribe", "https://www.kottke.org",
		)
	fmt.Printf(">> feeder -config test-data/subscribe-cfg.yml -subscribe https://www.kottke.org/ stdout:\n%s\n", stdOut)
	fmt.Printf(">> feeder -config test-data/subscribe-cfg.yml -subscribe https://www.kottke.org/ stderr:\n%s\n", stdErr)
	require.Zero(t, status)

	fs, err = readFeedsConfig("test-data/feeds.yml")
	require.Nil(t, err)
	require.Len(t, fs, 2)

	expected1 := &ConfigFeed{Name: "kottke.org - home of fine hypertext products", URL: "http://feeds.kottke.org/main"}
	require.Equal(t, expected0, fs[0])
	require.Equal(t, expected1, fs[1])
}
