package mysql_test

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
)

// SQLMock methods only used to mock rows, for everything else use the
// same as we use in all the app, Testify with Mockery.
func sqlmockRowsToStdRow(mRows *sqlmock.Rows) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("select").WillReturnRows(mRows)
	return db.QueryRow("select")
}

func sqlRowErr(err error) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("select").WillReturnError(err)
	return db.QueryRow("select")
}
