package mailing

import (
	"bytes"
	"fmt"
	"net/mail"

	mailutil "github.com/go-mail/mail"
)

func BuildMessage(from, to *mail.Address, subject, html, text string) ([]byte, error) {
	m := mailutil.NewMessage()
	m.SetHeader("From", from.String())
	m.SetHeader("To", to.String())
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", html)
	m.AddAlternative("text/plain", text)

	buff := &bytes.Buffer{}
	_, err := m.WriteTo(buff)
	if err != nil {
		return nil, fmt.Errorf("could not build mail body: %w", err)
	}

	return buff.Bytes(), nil
}
