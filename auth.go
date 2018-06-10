package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/lib/pq"
)

var (
	keyAuthUserID = ContextKey{"auth_user_id"}
	magicLinkTmpl = template.Must(template.ParseFiles("templates/magic-link.html"))
)

func passwordlessStart(w http.ResponseWriter, r *http.Request) {
	// Request parsing
	var input struct {
		Email       string `json:"email"`
		RedirectURI string `json:"redirectUri"`
	}
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
		respondJSON(w, Errors{errs}, http.StatusUnprocessableEntity)
		return
	}

	// Verification code
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

	// Magic link
	q := make(url.Values)
	q.Set("verification_code", verificationCode)
	q.Set("redirect_uri", input.RedirectURI)
	magicLink := config.domain
	magicLink.Path = "/api/passwordless/verify_redirect"
	magicLink.RawQuery = q.Encode()

	// Mailing
	var body bytes.Buffer
	data := map[string]string{"MagicLink": magicLink.String()}
	if err := magicLinkTmpl.Execute(&body, data); err != nil {
		respondInternalError(w, fmt.Errorf("could not execute magic link template: %v", err))
		return
	}
	if err := sendMail(input.Email, "Magic Link", body.String()); err != nil {
		respondInternalError(w, fmt.Errorf("could not mail magic link: %v", err))
		return
	}

	// Respond
	w.WriteHeader(http.StatusNoContent)
}

func passwordlessVerifyRedirect(w http.ResponseWriter, r *http.Request) {
	// Request parsing
	q := r.URL.Query()
	verificationCode := q.Get("verification_code")
	redirectURI := q.Get("redirect_uri")

	// Input validation
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
		respondJSON(w, Errors{errs}, http.StatusUnprocessableEntity)
		return
	}

	// Verification code
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

	// JWT
	expiresAt := time.Now().Add(time.Hour * 24 * 60) // 60 days
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   userID,
		ExpiresAt: expiresAt.Unix(),
	}).SignedString(config.jwtKey)
	if err != nil {
		respondInternalError(w, fmt.Errorf("could not create JWT: %v", err))
		return
	}

	// Callback
	expiresAtB, err := expiresAt.MarshalText()
	if err != nil {
		respondInternalError(w, fmt.Errorf("could not marshal expiration date: %v", err))
		return
	}
	f := make(url.Values)
	f.Set("jwt", tokenString)
	f.Set("expires_at", string(expiresAtB))
	callback.Fragment = f.Encode()

	// Redirect
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

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		if a := r.Header.Get("Authorization"); strings.HasPrefix(a, "Bearer ") {
			tokenString = a[7:]
		} else {
			next(w, r)
			return
		}

		p := jwt.Parser{ValidMethods: []string{jwt.SigningMethodHS256.Name}}
		token, err := p.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(*jwt.Token) (interface{}, error) { return config.jwtKey, nil })
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

func guard(handler, fallback http.HandlerFunc) http.HandlerFunc {
	return withAuth(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Value(keyAuthUserID).(string); !ok {
			if fallback != nil {
				fallback(w, r)
			} else {
				respondJSON(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			}
			return
		}
		handler(w, r)
	})
}
