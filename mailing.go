package main

import (
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
)

// MailSender function to send mails.
type MailSender func(to, subject, body string) error

func newMailSender(host string, port int, username, password string) MailSender {
	smtpAddr := net.JoinHostPort(host, strconv.Itoa(port))
	smtpAuth := smtp.PlainAuth("", username, password, host)
	from := mail.Address{
		Name:    "Passwordless Demo",
		Address: "noreply@" + origin.Hostname(),
	}
	return func(to, subject, body string) error {
		toAddr := mail.Address{Address: to}
		headers := map[string]string{
			"From":         from.String(),
			"To":           toAddr.String(),
			"Subject":      subject,
			"MIME-Version": "1.0",
			"Content-Type": `text/html; charset="utf-8"`,
		}
		msg := ""
		for k, v := range headers {
			msg += fmt.Sprintf("%s: %s\r\n", k, v)
		}
		msg += "\r\n"
		msg += body

		return smtp.SendMail(
			smtpAddr,
			smtpAuth,
			from.Address,
			[]string{toAddr.Address},
			[]byte(msg))
	}
}
