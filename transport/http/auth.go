package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nicolasparada/go-passwordless-demo"
)

func (h *handler) withAuthUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			authUserID, err := h.service.ParseAuthToken(auth[7:])
			if err != nil {
				h.respondErr(w, err)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, passwordless.KeyAuthUserID, authUserID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

type sendMagicLinkReqBody struct {
	Email       string
	RedirectURI string
}

func (h *handler) sendMagicLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var reqBody sendMagicLinkReqBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		h.respondErr(w, errBadRequest)
		return
	}

	ctx := r.Context()
	err = h.service.SendMagicLink(ctx, reqBody.Email, reqBody.RedirectURI)
	if err != nil {
		h.respondErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) verifyMagicLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	redirectURI, err := h.service.ValidateRedirectURI(q.Get("redirect_uri"))
	if err != nil {
		h.respondErr(w, err)
		return
	}

	email := q.Get("email")
	code := q.Get("code")
	username := emptyStringPtr(strings.TrimSpace(q.Get("username")))

	ctx := r.Context()
	auth, err := h.service.VerifyMagicLink(ctx, email, code, username)
	isRetryableError := err == passwordless.ErrUserNotFound ||
		err == passwordless.ErrInvalidUsername ||
		err == passwordless.ErrUsernameTaken
	if isRetryableError {
		h.redirectWithData(w, r, redirectURI, url.Values{
			"error":     []string{err.Error()},
			"retry_uri": []string{r.RequestURI},
		})
		return
	}
	if err != nil {
		h.redirectWithErr(w, r, redirectURI, err)
		return
	}

	h.redirectWithData(w, r, redirectURI, url.Values{
		"token":         []string{auth.Token},
		"expires_at":    []string{auth.ExpiresAt.Format(time.RFC3339Nano)},
		"user.id":       []string{auth.User.ID},
		"user.email":    []string{auth.User.Email},
		"user.username": []string{auth.User.Username},
	})
}

func (h *handler) authUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, err := h.service.AuthUser(ctx)
	if err != nil {
		h.respondErr(w, err)
		return
	}

	h.respond(w, u, http.StatusOK)
}

func emptyStringPtr(s string) *string {
	if s != "" {
		return &s
	}

	return nil
}
