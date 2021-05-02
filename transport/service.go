package transport

import (
	"context"
	"net/url"

	passwordless "github.com/nicolasparada/go-passwordless-demo"
)

type Service interface {
	SendMagicLink(ctx context.Context, email, redirectURI string) error
	ValidateRedirectURI(rawurl string) (*url.URL, error)
	VerifyMagicLink(ctx context.Context, email, code string, username *string) (passwordless.Auth, error)
	ParseAuthToken(token string) (userID string, err error)
	AuthUser(ctx context.Context) (passwordless.User, error)
}
