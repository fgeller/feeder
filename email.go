package main

import (
	"gopkg.in/gomail.v2"
)

func email(host string, port int, from, pass, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", from)
	m.SetHeader("Subject", "feeder says hello")
	m.SetBody("text/html", body)

	d := gomail.NewDialer(host, port, from, pass)
	return d.DialAndSend(m)
}
