package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	passwordless "github.com/nicolasparada/go-passwordless-demo"
)

func (repo *Repository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool

	query := "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)"
	row := repo.ext(ctx).QueryRowContext(ctx, query, email)
	err := row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("could not sql query select or scan user existence by email: %w", err)
	}

	return exists, nil
}

func (repo *Repository) UserByEmail(ctx context.Context, email string) (passwordless.User, error) {
	var u passwordless.User
	query := "SELECT id, username FROM users WHERE email = $1"
	row := repo.ext(ctx).QueryRowContext(ctx, query, email)
	err := row.Scan(&u.ID, &u.Username)
	if err == sql.ErrNoRows {
		return u, passwordless.ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("could not sql query select or scan user by email: %w", err)
	}

	u.Email = email

	return u, nil
}

func (repo *Repository) StoreUser(ctx context.Context, email, username string) (passwordless.User, error) {
	var u passwordless.User
	query := "INSERT INTO users (email, username) VALUES ($1, $2) RETURNING id"
	row := repo.ext(ctx).QueryRowContext(ctx, query, email, username)
	err := row.Scan(&u.ID)
	if isUniqueViolationError(err) {
		if strings.Contains(err.Error(), "email") {
			return u, passwordless.ErrEmailTaken
		}

		if strings.Contains(err.Error(), "username") {
			return u, passwordless.ErrUsernameTaken
		}
	}
	if err != nil {
		return u, fmt.Errorf("could not sql insert or scan user: %w", err)
	}

	u.Email = email
	u.Username = username

	return u, nil
}

func (repo *Repository) User(ctx context.Context, userID string) (passwordless.User, error) {
	var u passwordless.User
	query := "SELECT email, username FROM users WHERE id = $1"
	row := repo.ext(ctx).QueryRowContext(ctx, query, userID)
	err := row.Scan(&u.Email, &u.Username)
	if err == sql.ErrNoRows {
		return u, passwordless.ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("could not sql query select or scan user: %w", err)
	}

	u.ID = userID

	return u, nil
}

func isUniqueViolationError(err error) bool {
	e, ok := err.(*pq.Error)
	return ok && e.Code == "23505"
}
