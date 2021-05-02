package smtp

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"sync"
	"time"

	mailutil "github.com/go-mail/mail"
	"github.com/hako/durafmt"
	"github.com/nicolasparada/go-passwordless-demo/notification"
)

type Sender struct {
	FromName    string
	FromAddress string
	Host        string
	Port        uint64
	Username    string
	Password    string
	ComposeFunc func(ctx context.Context, to string, w io.Writer, data interface{}) error

	once sync.Once

	addr     string
	auth     smtp.Auth
	fromAddr *mail.Address
}

func (s *Sender) Send(ctx context.Context, data notification.MagicLinkData, to string) error {
	s.once.Do(func() {
		s.addr = net.JoinHostPort(s.Host, strconv.FormatUint(s.Port, 10))
		s.auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
		s.fromAddr = &mail.Address{Name: s.FromName, Address: s.FromAddress}
	})

	msg := &bytes.Buffer{}
	err := s.ComposeFunc(ctx, to, msg, data)
	if err != nil {
		return fmt.Errorf("could not compose magic link message: %w", err)
	}

	err = smtp.SendMail(s.addr, s.auth, s.fromAddr.String(), []string{to}, msg.Bytes())
	if err != nil {
		return fmt.Errorf("could not smtp send magic link: %w", err)
	}

	return nil
}

func buildMessage(w io.Writer, from, to *mail.Address, subject, html, text string) error {
	m := mailutil.NewMessage()
	m.SetHeader("From", from.String())
	m.SetHeader("To", to.String())
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", html)
	m.AddAlternative("text/plain", text)

	_, err := m.WriteTo(w)
	if err != nil {
		return fmt.Errorf("could not build mail body: %w", err)
	}

	return nil
}

var tmplFuncs = template.FuncMap{
	"human_duration": func(d time.Duration) string {
		return durafmt.Parse(d).LimitFirstN(1).String()
	},
}
