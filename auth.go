package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"
)

const (
	keyAuthUserID            = key("auth_user_id")
	verificationCodeLifespan = time.Minute * 15
	tokenLifespan            = time.Hour * 24 * 14
)

var magicLinkTmpl = template.Must(template.ParseFiles("templates/magic-link.html"))

type key string

func passwordlessStart(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email       string `json:"email"`
		RedirectURI string `json:"redirectUri"`
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respond(w, err.Error(), http.StatusBadRequest)
		return
	}

	errs := make(map[string]string)
	input.Email = strings.TrimSpace(input.Email)
	if input.Email == "" {
		errs["email"] = "Email required"
	} else if !rxEmail.MatchString(input.Email) {
		errs["email"] = "Invalid email"
	}
	input.RedirectURI = strings.TrimSpace(input.RedirectURI)
	if input.RedirectURI == "" {
		errs["redirectUri"] = "Redirect URI required"
	} else if _, err := url.ParseRequestURI(input.RedirectURI); err != nil {
		errs["redirectUri"] = "Invalid redirect URI"
	}
	if len(errs) != 0 {
		respond(w, Errors{errs}, http.StatusUnprocessableEntity)
		return
	}

	var verificationCode string
	err := db.QueryRowContext(r.Context(), `
		INSERT INTO verification_codes (user_id) VALUES
			((SELECT id FROM users WHERE email = $1))
		RETURNING id`, input.Email).Scan(&verificationCode)
	if isNotNullViolation(err) {
		respond(w, "No user found with that email", http.StatusNotFound)
		return
	} else if err != nil {
		respondErr(w, fmt.Errorf("could not insert verification code: %v", err))
		return
	}

	magicLink := *origin
	magicLink.Path = "/api/passwordless/verify_redirect"
	q := magicLink.Query()
	q.Set("verification_code", verificationCode)
	q.Set("redirect_uri", input.RedirectURI)
	magicLink.RawQuery = q.Encode()

	var b bytes.Buffer
	data := map[string]interface{}{
		"MagicLink": magicLink.String(),
		"Minutes":   int(verificationCodeLifespan.Minutes()),
	}
	if err := magicLinkTmpl.Execute(&b, data); err != nil {
		respondErr(w, fmt.Errorf("could not execute magic link template: %v", err))
		return
	}

	if err := mailSender.send(Mail{
		To:      mail.Address{Address: input.Email},
		Subject: "Magic Link",
		Body:    b.String(),
	}); err != nil {
		respondErr(w, fmt.Errorf("could not mail magic link: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func passwordlessVerifyRedirect(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	verificationCode := q.Get("verification_code")
	redirectURI := q.Get("redirect_uri")

	errs := make(map[string]string)
	verificationCode = strings.TrimSpace(verificationCode)
	if verificationCode == "" {
		errs["verification_code"] = "Verification code required"
	} else if !rxUUID.MatchString(verificationCode) {
		errs["verification_code"] = "Invalid verification code"
	}
	var callback *url.URL
	var err error
	redirectURI = strings.TrimSpace(redirectURI)
	if redirectURI == "" {
		errs["redirect_uri"] = "Redirect URI required"
	} else if callback, err = url.ParseRequestURI(redirectURI); err != nil {
		errs["redirect_uri"] = "Invalid redirect URI"
	}
	if len(errs) != 0 {
		respond(w, Errors{errs}, http.StatusUnprocessableEntity)
		return
	}

	var userID string
	var createdAt time.Time
	if err := db.QueryRowContext(r.Context(), `
		DELETE FROM verification_codes WHERE id = $1
		RETURNING user_id, created_at`, verificationCode).
		Scan(&userID, &createdAt); err == sql.ErrNoRows {
		respond(w, "Magic link not found", http.StatusNotFound)
		return
	} else if err != nil {
		respondErr(w, fmt.Errorf("could not delete verification code: %v", err))
		return
	}

	if createdAt.Add(verificationCodeLifespan).Before(time.Now()) {
		respond(w, "Link expired", http.StatusGone)
		return
	}

	token, err := codec.EncodeToString(userID)
	if err != nil {
		respondErr(w, fmt.Errorf("could not create token: %v", err))
		return
	}

	expiresAt, _ := time.Now().Add(tokenLifespan).MarshalText()

	f := url.Values{}
	f.Set("token", token)
	f.Set("expires_at", string(expiresAt))
	callback.Fragment = f.Encode()

	http.Redirect(w, r, callback.String(), http.StatusFound)
}

func getAuthUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authUserID := ctx.Value(keyAuthUserID).(string)

	user, err := userByID(ctx, authUserID)
	if err == sql.ErrNoRows {
		respond(w, http.StatusText(http.StatusTeapot), http.StatusTeapot)
		return
	} else if err != nil {
		respondErr(w, fmt.Errorf("could not query auth user: %v", err))
		return
	}

	respond(w, user, http.StatusOK)
}

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			next(w, r)
			return
		}

		token := authHeader[7:]
		userID, err := codec.DecodeToString(token)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		userID = strings.TrimSpace(userID)
		if !rxUUID.MatchString(userID) {
			http.Error(w, http.StatusText(http.StatusTeapot), http.StatusTeapot)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, keyAuthUserID, userID)

		next(w, r.WithContext(ctx))
	}
}

func guard(next http.HandlerFunc) http.HandlerFunc {
	return withAuth(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Value(keyAuthUserID).(string); !ok {
			respond(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next(w, r)
	})
}
