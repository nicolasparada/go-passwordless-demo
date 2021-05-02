package cockroach

import (
	"context"
	"database/sql"
	"fmt"

	passwordless "github.com/nicolasparada/go-passwordless-demo"
)

func (repo *Repository) StoreVerificationCode(ctx context.Context, email string) (passwordless.VerificationCode, error) {
	var vc passwordless.VerificationCode

	query := "INSERT INTO verification_codes (email) VALUES ($1) RETURNING code, created_at"
	row := repo.ext(ctx).QueryRowContext(ctx, query, email)
	err := row.Scan(&vc.Code, &vc.CreatedAt)
	if err != nil {
		return vc, fmt.Errorf("could not sql insert or scan verification code: %w", err)
	}

	return vc, nil
}

func (repo *Repository) VerificationCode(ctx context.Context, email, code string) (passwordless.VerificationCode, error) {
	var data passwordless.VerificationCode

	query := "SELECT created_at FROM verification_codes WHERE email = $1 AND code = $2"
	row := repo.ext(ctx).QueryRowContext(ctx, query, email, code)
	err := row.Scan(&data.CreatedAt)
	if err == sql.ErrNoRows {
		return data, passwordless.ErrVerificationCodeNotFound
	}

	if err != nil {
		return data, fmt.Errorf("could not sql query select or scan verification code: %w", err)
	}

	data.Email = email
	data.Code = code

	return data, nil
}

func (repo *Repository) DeleteVerificationCode(ctx context.Context, email, code string) (bool, error) {
	query := "DELETE FROM verification_codes WHERE email = $1 AND code = $2"
	result, err := repo.ext(ctx).ExecContext(ctx, query, email, code)
	if err != nil {
		return false, fmt.Errorf("could not sql delete verification code: %w", err)
	}

	ra, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("could not sql count deleted verification code rows: %w", err)
	}

	return ra != 0, nil
}
