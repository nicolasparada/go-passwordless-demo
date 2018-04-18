package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/lib/pq"
)

// User represents an authenticated user or a resource owner.
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// CreateUserRequest holds the JSON request body.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

var (
	rxEmail    = regexp.MustCompile("^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$")
	rxUsername = regexp.MustCompile("^[a-zA-Z][\\w|-]{1,17}$")
)

func createUser(w http.ResponseWriter, r *http.Request) {
	// Request parsing
	var input CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondJSON(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Input validation
	errs := make(map[string]string)
	if input.Email == "" {
		errs["email"] = "Email required"
	} else if !rxEmail.MatchString(input.Email) {
		errs["email"] = "Invalid email"
	}
	if input.Username == "" {
		errs["username"] = "Username required"
	}
	if len(errs) != 0 {
		respondJSON(w, errs, http.StatusUnprocessableEntity)
		return
	}

	// Insert user
	var user User
	err := db.QueryRowContext(r.Context(), `
		INSERT INTO users (email, username) VALUES ($1, $2)
		RETURNING id
	`, input.Email, input.Username).Scan(&user.ID)
	if errPq, ok := err.(*pq.Error); ok && errPq.Code.Name() == "unique_violation" {
		if strings.Contains(errPq.Error(), "email") {
			errs["email"] = "Email taken"
		} else {
			errs["username"] = "Username taken"
		}
		respondJSON(w, errs, http.StatusForbidden)
		return
	} else if err != nil {
		respondInternalError(w, fmt.Errorf("could not insert user: %v", err))
		return
	}

	user.Email = input.Email
	user.Username = input.Username

	respondJSON(w, user, http.StatusCreated)
}

func fetchUser(ctx context.Context, id string) (User, error) {
	user := User{ID: id}
	err := db.QueryRowContext(ctx, `
		SELECT email, username FROM users WHERE id = $1
	`, id).Scan(&user.Email, &user.Username)
	return user, err
}
