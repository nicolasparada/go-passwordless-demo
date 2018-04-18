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
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/lib/pq"
)

// PasswordlessStartRequest holds the JSON request body.
type PasswordlessStartRequest struct {
	Email       string `json:"email"`
	RedirectURI string `json:"redirectUri"`
}

const (
	keyAuthUserID ContextKey = iota
)

var (
	rxUUID        = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")
	magicLinkTmpl = template.Must(template.ParseFiles("templates/magic-link.html"))
)

func passwordlessStart(w http.ResponseWriter, r *http.Request) {
	// Request parsing
	var input PasswordlessStartRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Input validation
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
		respondJSON(w, errs, http.StatusUnprocessableEntity)
		return
	}

	// Insert verification code
	var verificationCode string
	err := db.QueryRowContext(r.Context(), `
		INSERT INTO verification_codes (user_id) VALUES
			((SELECT id FROM users WHERE email = $1))
		RETURNING id
	`, input.Email).Scan(&verificationCode)
	if errPq, ok := err.(*pq.Error); ok && errPq.Code.Name() == "not_null_violation" {
		respondJSON(w, "No user found with that email", http.StatusNotFound)
		return
	} else if err != nil {
		respondInternalError(w, fmt.Errorf("could not insert verification code: %v", err))
		return
	}

	// Create magic link
	q := make(url.Values)
	q.Set("verification_code", verificationCode)
	q.Set("redirect_uri", input.RedirectURI)
	magicLink := *config.appURL
	magicLink.Path = "/api/passwordless/verify_redirect"
	magicLink.RawQuery = q.Encode()

	// Creating mail
	var body bytes.Buffer
	data := map[string]string{"MagicLink": magicLink.String()}
	if err := magicLinkTmpl.Execute(&body, data); err != nil {
		respondInternalError(w, fmt.Errorf("could not execute magic link template: %v", err))
		return
	}

	// Sending mail
	to := mail.Address{Address: input.Email}
	if err := sendMail(to, "Magic Link", body.String()); err != nil {
		respondInternalError(w, fmt.Errorf("could not mail magic link: %v", err))
		return
	}

	// Respond OK
	w.WriteHeader(http.StatusNoContent)
}

func passwordlessVerifyRedirect(w http.ResponseWriter, r *http.Request) {
	// Parse request input
	q := r.URL.Query()
	verificationCode := q.Get("verification_code")
	redirectURI := q.Get("redirect_uri")

	// Validate input
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
		respondJSON(w, errs, http.StatusUnprocessableEntity)
		return
	}

	// Delete verification code
	var userID string
	if err := db.QueryRowContext(r.Context(), `
		DELETE FROM verification_codes
		WHERE id = $1
			AND created_at >= now() - INTERVAL '15m'
		RETURNING user_id
	`, verificationCode).Scan(&userID); err == sql.ErrNoRows {
		respondJSON(w, "Link expired or already used", http.StatusBadRequest)
		return
	} else if err != nil {
		respondInternalError(w, fmt.Errorf("could not delete verification code: %v", err))
		return
	}

	// Create JWT
	expiresAt := time.Now().Add(time.Hour * 24 * 60) // 60 days
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   userID,
		ExpiresAt: expiresAt.Unix(),
	}).SignedString(config.jwtKey)
	if err != nil {
		respondInternalError(w, fmt.Errorf("could not create JWT: %v", err))
		return
	}

	// Redirect to callback URL
	expiresAtB, err := expiresAt.MarshalText()
	if err != nil {
		respondInternalError(w, fmt.Errorf("could not marshal expiration date: %v", err))
		return
	}
	f := make(url.Values)
	f.Set("jwt", tokenString)
	f.Set("expires_at", string(expiresAtB))
	callback.Fragment = f.Encode()

	http.Redirect(w, r, callback.String(), http.StatusFound)
}

func getAuthUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authUserID := ctx.Value(keyAuthUserID).(string)

	user, err := fetchUser(ctx, authUserID)
	if err == sql.ErrNoRows {
		respondJSON(w, http.StatusText(http.StatusTeapot), http.StatusTeapot)
		return
	} else if err != nil {
		respondInternalError(w, fmt.Errorf("could not query auth user: %v", err))
		return
	}

	respondJSON(w, user, http.StatusOK)
}

func jwtKeyFunc(*jwt.Token) (interface{}, error) { return config.jwtKey, nil }

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := r.Header.Get("Authorization")
		hasToken := strings.HasPrefix(a, "Bearer ")
		if !hasToken {
			next(w, r)
			return
		}
		tokenString := a[7:]

		p := jwt.Parser{ValidMethods: []string{jwt.SigningMethodHS256.Name}}
		token, err := p.ParseWithClaims(tokenString, &jwt.StandardClaims{}, jwtKeyFunc)
		if err != nil {
			respondJSON(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*jwt.StandardClaims)
		if !ok || !token.Valid {
			respondJSON(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, keyAuthUserID, claims.Subject)

		next(w, r.WithContext(ctx))
	}
}

func authRequired(next http.HandlerFunc) http.HandlerFunc {
	return withAuth(func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value(keyAuthUserID).(string)
		if !ok {
			respondJSON(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next(w, r)
	})
}
