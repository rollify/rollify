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

func sqlmockRowsToStdRows(mRows *sqlmock.Rows) *sql.Rows {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("select").WillReturnRows(mRows)
	r, _ := db.Query("select")
	return r
}

func sqlRowErr(err error) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("select").WillReturnError(err)
	return db.QueryRow("select")
}
