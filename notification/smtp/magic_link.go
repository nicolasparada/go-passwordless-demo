package smtp

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/mail"

	"github.com/nicolasparada/go-passwordless-demo/notification"
	"github.com/nicolasparada/go-passwordless-demo/web"
	"golang.org/x/sync/errgroup"
)

// MagicLinkComposer handles web/template/mail/magic-link.{html,txt}.tmpl composing.
// Uses notification.MagicLinkData as data.
func MagicLinkComposer(fromName, fromAddr string) (notification.ComposeFunc, error) {
	b, err := web.Files.ReadFile("template/mail/magic-link.html.tmpl")
	if err != nil {
		return nil, fmt.Errorf("could not read magic link html template file: %w", err)
	}

	htmlTmpl, err := template.New("mail/magic-link.html").Funcs(tmplFuncs).Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("could not parse magic link html template: %w", err)
	}

	b, err = web.Files.ReadFile("template/mail/magic-link.txt.tmpl")
	if err != nil {
		return nil, fmt.Errorf("could not read magic link plain text template file: %w", err)
	}

	plainTextTmpl, err := template.New("mail/magic-link.txt").Funcs(tmplFuncs).Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("could not parse magic link plain text template: %w", err)
	}

	from := &mail.Address{Name: fromName, Address: fromAddr}

	composeFunc := func(ctx context.Context, email string, w io.Writer, v interface{}) error {
		data, ok := v.(notification.MagicLinkData)
		if !ok {
			return fmt.Errorf("unexpected magic link data type %T", v)
		}

		htmlRenderer, plainTextRenderer := &bytes.Buffer{}, &bytes.Buffer{}
		g := &errgroup.Group{}
		g.Go(func() error {
			err := htmlTmpl.Execute(htmlRenderer, data)
			if err != nil {
				return fmt.Errorf("could not render magic link html template: %w", err)
			}

			return nil
		})
		g.Go(func() error {
			err := plainTextTmpl.Execute(plainTextRenderer, data)
			if err != nil {
				return fmt.Errorf("could not render magic link plain text template: %w", err)
			}

			return nil
		})

		if err := g.Wait(); err != nil {
			return err
		}

		to := &mail.Address{Address: email}
		subject := "Login to Golang Passwordless Demo"
		html := htmlRenderer.String()
		plainText := plainTextRenderer.String()
		err = buildMessage(w, from, to, subject, html, plainText)
		if err != nil {
			return err
		}

		return nil
	}

	return composeFunc, nil
}
