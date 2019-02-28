package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"runtime/debug"
)

// SPAFileSystem file system with single-page applications support.
type SPAFileSystem struct {
	root http.Dir
}

// Open wraps http.Dir Open method to enable single-page applications.
func (fs SPAFileSystem) Open(name string) (http.File, error) {
	f, err := fs.root.Open(name)
	if os.IsNotExist(err) {
		return fs.root.Open("index.html")
	}
	return f, err
}

func respond(w http.ResponseWriter, payload interface{}, statusCode int) {
	b, err := json.Marshal(payload)
	if err != nil {
		respondErr(w, fmt.Errorf("could not marshal response payload: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(b)
}

func respondErr(w http.ResponseWriter, err error) {
	log.Println(err)
	respond(w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func requireJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || ct != "application/json" {
			respond(w, "content type of application/json required", http.StatusUnsupportedMediaType)
			return
		}
		next(w, r)
	}
}

func withRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic: %+v\n", r)
				debug.PrintStack()

				var err error
				switch x := r.(type) {
				case error:
					err = x
				case string:
					err = errors.New(x)
				default:
					err = errors.New("Unknown panic")
				}
				respondErr(w, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
