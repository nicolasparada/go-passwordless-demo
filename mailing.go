package main

import (
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
)

var sendMail func(to, subject, body string) error

func initMailing() {
	smtpAddr := net.JoinHostPort(config.smtpHost, strconv.Itoa(config.smtpPort))
	smtpAuth := smtp.PlainAuth("", config.smtpUsername, config.smtpPassword, config.smtpHost)
	from := mail.Address{
		Name:    "Passwordless Demo",
		Address: "noreply@" + config.domain.Host,
	}
	sendMail = func(to, subject, body string) error {
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
