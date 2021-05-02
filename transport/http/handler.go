package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	passwordless "github.com/nicolasparada/go-passwordless-demo"
	"github.com/nicolasparada/go-passwordless-demo/transport"
)

var errBadRequest = errors.New("bad request")

func NewHandler(svc transport.Service, l *log.Logger) http.Handler {
	h := &handler{service: svc, logger: l}
	api := http.NewServeMux()
	api.HandleFunc("/api/send-magic-link", h.sendMagicLink)
	api.HandleFunc("/api/verify-magic-link", h.verifyMagicLink)
	api.HandleFunc("/api/auth-user", h.authUser)

	mux := http.NewServeMux()
	mux.Handle("/api/", h.withAuthUserID(api))
	mux.Handle("/", staticHandler())
	return mux
}

type handler struct {
	service transport.Service
	logger  *log.Logger
}

func (h *handler) respond(w http.ResponseWriter, v interface{}, statusCode int) {
	b, err := json.Marshal(v)
	if err != nil {
		h.respondErr(w, fmt.Errorf("could not json marshall http response body: %w", err))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_, err = w.Write(b)
	if err != nil && !errors.Is(err, context.Canceled) {
		h.logger.Printf("could not write http response: %v\n", err)
	}
}

func (h *handler) respondErr(w http.ResponseWriter, err error) {
	statusCode := err2code(err)
	if statusCode != http.StatusInternalServerError {
		http.Error(w, err.Error(), statusCode)
		return
	}

	h.logger.Println(err)
	http.Error(w, "internal server error", statusCode)
}

func (h *handler) redirectWithErr(w http.ResponseWriter, r *http.Request, uri *url.URL, err error) {
	statusCode := err2code(err)
	if statusCode != http.StatusInternalServerError {
		h.redirectWithData(w, r, uri, url.Values{"error": []string{err.Error()}})
		return
	}

	h.logger.Println(err)
	h.redirectWithData(w, r, uri, url.Values{"error": []string{"internal server error"}})
}

func (h *handler) redirectWithData(w http.ResponseWriter, r *http.Request, uri *url.URL, data url.Values) {
	// Initially using query string instead of hash fragment
	// and replacing "?" by "#" later
	// because golang's RawFragment is a no-op.
	uri.RawQuery = data.Encode()
	location := uri.String()
	location = strings.Replace(location, "?", "#", 1)
	http.Redirect(w, r, location, http.StatusFound)
}

func err2code(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch err {
	case errBadRequest:
		return http.StatusBadRequest
	case passwordless.ErrInvalidEmail,
		passwordless.ErrInvalidRedirectURI,
		passwordless.ErrInvalidVerificationCode,
		passwordless.ErrInvalidUsername:
		return http.StatusUnprocessableEntity
	case passwordless.ErrUntrustedRedirectURI:
		return http.StatusForbidden
	case passwordless.ErrVerificationCodeNotFound,
		passwordless.ErrUserNotFound:
		return http.StatusNotFound
	case passwordless.ErrVerificationCodeExpired:
		return http.StatusUnauthorized
	case passwordless.ErrEmailTaken,
		passwordless.ErrUsernameTaken:
		return http.StatusConflict
	case passwordless.ErrUnauthenticated:
		return http.StatusUnauthorized
	}

	return http.StatusInternalServerError
}
