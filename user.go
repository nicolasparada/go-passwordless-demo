package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// User represents an authenticated user or a resource owner.
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Username string `json:"username"`
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
	input.Username = strings.TrimSpace(input.Username)
	if input.Username == "" {
		errs["username"] = "Username required"
	} else if !rxUsername.MatchString(input.Username) {
		errs["username"] = "Invalid username"
	}
	if len(errs) != 0 {
		respond(w, Errors{errs}, http.StatusUnprocessableEntity)
		return
	}

	var user User
	err := db.QueryRowContext(r.Context(), `
		INSERT INTO users (email, username) VALUES ($1, $2)
		RETURNING id`, input.Email, input.Username).Scan(&user.ID)
	if isUniqueViolation(err) {
		if strings.Contains(err.Error(), "email") {
			errs["email"] = "Email taken"
		} else {
			errs["username"] = "Username taken"
		}
		respond(w, errs, http.StatusForbidden)
		return
	} else if err != nil {
		respondErr(w, fmt.Errorf("could not insert user: %v", err))
		return
	}

	user.Email = input.Email
	user.Username = input.Username

	respond(w, user, http.StatusCreated)
}

func userByID(ctx context.Context, id string) (User, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var user User
	if err := db.QueryRowContext(ctx, "SELECT email, username FROM users WHERE id = $1", id).
		Scan(&user.Email, &user.Username); err != nil {
		return user, err
	}

	user.ID = id
	return user, nil
}
