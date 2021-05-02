package smtp

import (
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"sync"

	"github.com/nicolasparada/go-passwordless-demo/mailing"
)

type Sender struct {
	FromName    string
	FromAddress string
	Host        string
	Port        uint64
	Username    string
	Password    string

	once sync.Once

	addr     string
	auth     smtp.Auth
	fromAddr *mail.Address
}

func (s *Sender) Send(to, subject, html, text string) error {
	s.once.Do(func() {
		s.addr = net.JoinHostPort(s.Host, strconv.FormatUint(s.Port, 10))
		s.auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
		s.fromAddr = &mail.Address{Name: s.FromName, Address: s.FromAddress}
	})

	msg, err := mailing.BuildMessage(s.fromAddr, &mail.Address{Address: to}, subject, html, text)
	if err != nil {
		return err
	}

	err = smtp.SendMail(s.addr, s.auth, s.fromAddr.String(), []string{to}, msg)
	if err != nil {
		return fmt.Errorf("could not smtp send mail: %w", err)
	}

	return nil
}
