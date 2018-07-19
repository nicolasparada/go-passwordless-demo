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
	"strconv"
	"strings"
	"time"

	"github.com/knq/jwt"
	"github.com/lib/pq"
)

const jwtLifespan = time.Hour * 24 * 14 // 14 days

var (
	keyAuthUserID = ContextKey{"auth_user_id"}
	magicLinkTmpl = template.Must(template.ParseFiles("templates/magic-link.html"))
)

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
	if input.Email == "" {
		errs["email"] = "Email required"
	} else if !rxEmail.MatchString(input.Email) {
		errs["email"] = "Invalid email"
	}
	if input.RedirectURI == "" {
		errs["redirectUri"] = "Redirect URI required"
	} else if u, err := url.Parse(input.RedirectURI); err != nil || !u.IsAbs() {
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
		RETURNING id
	`, input.Email).Scan(&verificationCode)
	if errPq, ok := err.(*pq.Error); ok && errPq.Code.Name() == "not_null_violation" {
		respond(w, "No user found with that email", http.StatusNotFound)
		return
	} else if err != nil {
		respondError(w, fmt.Errorf("could not insert verification code: %v", err))
		return
	}

	q := make(url.Values)
	q.Set("verification_code", verificationCode)
	q.Set("redirect_uri", input.RedirectURI)
	magicLink := *origin
	magicLink.Path = "/api/passwordless/verify_redirect"
	magicLink.RawQuery = q.Encode()

	var body bytes.Buffer
	data := map[string]string{"MagicLink": magicLink.String()}
	if err := magicLinkTmpl.Execute(&body, data); err != nil {
		respondError(w, fmt.Errorf("could not execute magic link template: %v", err))
		return
	}

	if err := mailSender.send(Mail{
		To:      mail.Address{Address: input.Email},
		Subject: "Magic Link",
		Body:    body.String(),
	}); err != nil {
		respondError(w, fmt.Errorf("could not mail magic link: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func passwordlessVerifyRedirect(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	verificationCode := q.Get("verification_code")
	redirectURI := q.Get("redirect_uri")

	errs := make(map[string]string)
	if verificationCode == "" {
		errs["verification_code"] = "Verification code required"
	} else if !rxUUID.MatchString(verificationCode) {
		errs["verification_code"] = "Invalid verification code"
	}
	var callback *url.URL
	var err error
	if redirectURI == "" {
		errs["redirect_uri"] = "Redirect URI required"
	} else if callback, err = url.Parse(redirectURI); err != nil || !callback.IsAbs() {
		errs["redirect_uri"] = "Invalid redirect URI"
	}
	if len(errs) != 0 {
		respond(w, Errors{errs}, http.StatusUnprocessableEntity)
		return
	}

	var userID string
	if err := db.QueryRowContext(r.Context(), `
		DELETE FROM verification_codes
		WHERE id = $1
			AND created_at >= now() - INTERVAL '15m'
		RETURNING user_id
	`, verificationCode).Scan(&userID); err == sql.ErrNoRows {
		respond(w, "Link expired or already used", http.StatusBadRequest)
		return
	} else if err != nil {
		respondError(w, fmt.Errorf("could not delete verification code: %v", err))
		return
	}

	exp := time.Now().Add(jwtLifespan)
	token, err := jwtSigner.Encode(jwt.Claims{
		Subject:    userID,
		Expiration: json.Number(strconv.FormatInt(exp.Unix(), 10)),
	})
	if err != nil {
		respondError(w, fmt.Errorf("could not create JWT: %v", err))
		return
	}

	expiresAt, _ := exp.MarshalText()
	f := make(url.Values)
	f.Set("token", string(token))
	f.Set("expires_at", string(expiresAt))
	callback.Fragment = f.Encode()

	http.Redirect(w, r, callback.String(), http.StatusFound)
}

func getAuthUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authUserID := ctx.Value(keyAuthUserID).(string)

	user, err := fetchUser(ctx, authUserID)
	if err == sql.ErrNoRows {
		respond(w, http.StatusText(http.StatusTeapot), http.StatusTeapot)
		return
	} else if err != nil {
		respondError(w, fmt.Errorf("could not query auth user: %v", err))
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

		var claims jwt.Claims
		if err := jwtSigner.Decode([]byte(token), &claims); err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, keyAuthUserID, claims.Subject)

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
