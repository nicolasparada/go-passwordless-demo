package main

import (
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
)

// MailSender provides a method to send mails.
type MailSender struct {
	Addr string
	Auth smtp.Auth
	From mail.Address
}

// Mail to send.
type Mail struct {
	To      mail.Address
	Subject string
	Body    string
}

func newMailSender(host string, port int, username, password string) *MailSender {
	return &MailSender{
		Addr: net.JoinHostPort(host, strconv.Itoa(port)),
		Auth: smtp.PlainAuth("", username, password, host),
		From: mail.Address{
			Name:    "Passwordless Demo",
			Address: "noreply@" + origin.Hostname(),
		},
	}
}

func (s *MailSender) send(mail Mail) error {
	headers := map[string]string{
		"From":         s.From.String(),
		"To":           mail.To.String(),
		"Subject":      mail.Subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=utf-8",
	}
	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n"
	msg += mail.Body

	return smtp.SendMail(
		s.Addr,
		s.Auth,
		s.From.Address,
		[]string{mail.To.Address},
		[]byte(msg))
}
