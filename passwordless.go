package passwordless

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"time"

	"github.com/hako/branca"
	"github.com/nicolasparada/go-passwordless-demo/notification"
)

const (
	verificationCodeTTL = time.Minute * 20
	authTokenTTL        = time.Hour * 24 * 14
)

var KeyAuthUserID = struct{ name string }{name: "key-auth-user-id"}

var (
	ErrInvalidEmail             = errors.New("invalid email")
	ErrInvalidRedirectURI       = errors.New("invalid redirect URI")
	ErrUntrustedRedirectURI     = errors.New("untrusted redirect URI")
	ErrInvalidVerificationCode  = errors.New("invalid verification code")
	ErrInvalidUsername          = errors.New("invalid username")
	ErrVerificationCodeNotFound = errors.New("verification code not found")
	ErrVerificationCodeExpired  = errors.New("verification code expired")
	ErrUserNotFound             = errors.New("user not found")
	ErrEmailTaken               = errors.New("email taken")
	ErrUsernameTaken            = errors.New("username taken")
	ErrUnauthenticated          = errors.New("unauthenticated")
)

type Service struct {
	Logger          *log.Logger
	Origin          *url.URL
	Repository      Repository
	MagicLinkSender NotificationSender
	AuthTokenKey    string
}

type Repository interface {
	ExecuteTx(ctx context.Context, txFunc func(ctx context.Context) error) error

	StoreVerificationCode(ctx context.Context, email string) (VerificationCode, error)
	VerificationCode(ctx context.Context, email, code string) (VerificationCode, error)
	DeleteVerificationCode(ctx context.Context, email, code string) (bool, error)

	UserExistsByEmail(ctx context.Context, email string) (bool, error)
	UserByEmail(ctx context.Context, email string) (User, error)
	StoreUser(ctx context.Context, email, username string) (User, error)
	User(ctx context.Context, userID string) (User, error)
}

type VerificationCode struct {
	Email     string
	Code      string
	CreatedAt time.Time
}

func (vc VerificationCode) Expired() bool {
	return vc.CreatedAt.Add(verificationCodeTTL).Before(time.Now())
}

type NotificationSender interface {
	Send(ctx context.Context, data notification.MagicLinkData, to string) error
}

type Auth struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	User      User      `json:"user"`
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (svc *Service) SendMagicLink(ctx context.Context, email, redirectURI string) error {
	if !isValidEmail(email) {
		return ErrInvalidEmail
	}

	_, err := svc.ValidateRedirectURI(redirectURI)
	if err != nil {
		return err
	}

	vc, err := svc.Repository.StoreVerificationCode(ctx, email)
	if err != nil {
		return err
	}

	// See transport/http/handler.go
	q := url.Values{}
	q.Set("email", email)
	q.Set("code", vc.Code)
	q.Set("redirect_uri", redirectURI)
	magicLink := cloneURL(svc.Origin)
	magicLink.Path = "/api/verify-magic-link"
	magicLink.RawQuery = q.Encode()

	err = svc.MagicLinkSender.Send(ctx, notification.MagicLinkData{
		Origin:    svc.Origin,
		TTL:       verificationCodeTTL,
		MagicLink: magicLink,
	}, email)
	if err != nil {
		return fmt.Errorf("could not send magic link to user: %w", err)
	}

	return nil
}

func (svc *Service) ValidateRedirectURI(rawurl string) (*url.URL, error) {
	uri, err := url.Parse(rawurl)
	if err != nil || !uri.IsAbs() {
		return nil, ErrInvalidRedirectURI
	}

	if uri.Host != svc.Origin.Host {
		return nil, ErrUntrustedRedirectURI
	}

	return uri, nil
}

func (svc *Service) VerifyMagicLink(ctx context.Context, email, code string, username *string) (Auth, error) {
	var auth Auth

	if !isValidEmail(email) {
		return auth, ErrInvalidEmail
	}

	if !isValidVerificationCode(code) {
		return auth, ErrInvalidVerificationCode
	}

	if username != nil && !isValidUsername(*username) {
		return auth, ErrInvalidUsername
	}

	vc, err := svc.Repository.VerificationCode(ctx, email, code)
	if err != nil {
		return auth, err
	}

	if vc.Expired() {
		return auth, ErrVerificationCodeExpired
	}

	err = svc.Repository.ExecuteTx(ctx, func(ctx context.Context) error {
		exists, err := svc.Repository.UserExistsByEmail(ctx, vc.Email)
		if err != nil {
			return err
		}

		if exists {
			auth.User, err = svc.Repository.UserByEmail(ctx, vc.Email)
			return err
		}

		if username == nil {
			return ErrUserNotFound
		}

		auth.User, err = svc.Repository.StoreUser(ctx, vc.Email, *username)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return auth, err
	}

	auth.ExpiresAt = time.Now().Add(authTokenTTL)
	auth.Token, err = svc.authTokenCodec().EncodeToString(auth.User.ID)
	if err != nil {
		return auth, fmt.Errorf("could not generate auth token: %w", err)
	}

	go func() {
		_, err := svc.Repository.DeleteVerificationCode(context.Background(), email, code)
		if err != nil {
			svc.Logger.Printf("failed to delete verification code: %v\n", err)
		}
	}()

	return auth, nil
}

func (svc *Service) authTokenCodec() *branca.Branca {
	cdc := branca.NewBranca(svc.AuthTokenKey)
	cdc.SetTTL(uint32(authTokenTTL.Seconds()))
	return cdc
}

func (svc *Service) ParseAuthToken(token string) (userID string, err error) {
	userID, err = svc.authTokenCodec().DecodeToString(token)
	if err == branca.ErrInvalidToken || err == branca.ErrInvalidTokenVersion {
		return "", ErrUnauthenticated
	}

	if _, ok := err.(*branca.ErrExpiredToken); ok {
		return "", ErrUnauthenticated
	}

	return userID, err
}

func (svc *Service) AuthUser(ctx context.Context) (User, error) {
	var u User

	authUserID, ok := ctx.Value(KeyAuthUserID).(string)
	if !ok {
		return u, ErrUnauthenticated
	}

	return svc.Repository.User(ctx, authUserID)
}

var reEmail = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func isValidEmail(s string) bool {
	return reEmail.MatchString(s)
}

var reUsername = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,17}$`)

func isValidUsername(s string) bool {
	return reUsername.MatchString(s)
}

var reUUID4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func isValidVerificationCode(s string) bool {
	return reUUID4.MatchString(s)
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	u2 := new(url.URL)
	*u2 = *u
	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}
	return u2
}
