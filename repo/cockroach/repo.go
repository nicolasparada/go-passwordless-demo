package cockroach

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cockroachdb/cockroach-go/crdb"
)

var keyTx = struct{ name string }{name: "key-tx"}

type Repository struct {
	DB                 *sql.DB
	DisableCRDBRetries bool
}

type ext interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func (repo *Repository) ext(ctx context.Context) ext {
	tx, ok := ctx.Value(keyTx).(*sql.Tx)
	if !ok {
		return repo.DB
	}

	return tx
}

func (repo *Repository) ExecuteTx(ctx context.Context, txFunc func(ctx context.Context) error) error {
	if repo.DisableCRDBRetries {
		tx, err := repo.DB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("could not begin tx: %w", err)
		}

		defer func() {
			_ = tx.Rollback()
		}()

		err = txFunc(context.WithValue(ctx, keyTx, tx))
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("could not commit tx: %w", err)
		}

		return nil
	}

	return crdb.ExecuteTx(ctx, repo.DB, nil, func(tx *sql.Tx) error {
		return txFunc(context.WithValue(ctx, keyTx, tx))
	})
}
