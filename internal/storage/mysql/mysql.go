package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

// DBClient is the Database client.
type DBClient interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

//go:generate mockery --case underscore --output mysqlmock --outpkg mysqlmock --name DBClient

func isDuplicateKeyError(err error) bool {
	const mysqlErrCode = 1062 // `ER_DUP_ENTRY`

	merr := &mysql.MySQLError{}
	if errors.As(err, &merr) {
		return merr.Number == mysqlErrCode
	}

	return false
}
