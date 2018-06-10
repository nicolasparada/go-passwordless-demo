package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

// ContextKey used on app middlewares.
type ContextKey struct {
	Name string
}

// SPAFileSystem file system with single-page applications support.
type SPAFileSystem struct {
	fs http.FileSystem
}

// Open wraps http.Dir Open method to enable single-page applications.
func (fs SPAFileSystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return fs.fs.Open("index.html")
	}
	return f, nil
}

func respondJSON(w http.ResponseWriter, payload interface{}, statusCode int) {
	switch value := payload.(type) {
	case string:
		payload = map[string]string{"message": value}
	case int:
		payload = map[string]int{"value": value}
	case bool:
		payload = map[string]bool{"result": value}
	}
	b, err := json.Marshal(payload)
	if err != nil {
		respondInternalError(w, fmt.Errorf("could not marshal response payload: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(b)
}

func respondInternalError(w http.ResponseWriter, err error) {
	log.Println(err)
	respondJSON(w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func requireJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		isJSON := strings.HasPrefix(ct, "application/json")
		if !isJSON {
			respondJSON(w, "JSON body required", http.StatusUnsupportedMediaType)
			return
		}
		next(w, r)
	}
}

func withRecover(next http.Handler) http.Handler {
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
				respondInternalError(w, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
