package mysql_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	drivermysql "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage/mysql"
	"github.com/rollify/rollify/internal/storage/mysql/mysqlmock"
)

func TestRoomRepositoryCreateRoom(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")
	wantedErr := fmt.Errorf("wanted error")

	tests := map[string]struct {
		config mysql.RoomRepositoryConfig
		mock   func(*mysqlmock.DBClient)
		room   model.Room
		expErr error
	}{
		"Having an error while storing the room, should error.": {
			config: mysql.RoomRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				m.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, wantedErr)
			},
			room: model.Room{
				ID:        "test-id",
				CreatedAt: t0,
				Name:      "test",
			},
			expErr: wantedErr,
		},

		"Creating the same room when already exists, should error.": {
			config: mysql.RoomRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				err := &drivermysql.MySQLError{Number: 1062}
				m.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, err)
			},
			room: model.Room{
				ID:        "test-id",
				CreatedAt: t0,
				Name:      "test",
			},
			expErr: internalerrors.ErrAlreadyExists,
		},

		"Creating a room should store the room.": {
			config: mysql.RoomRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "INSERT INTO room (id, name, created_at) VALUES (?, ?, ?)"
				m.On("ExecContext", mock.Anything, expQuery, "test-id", "test", t0).Once().Return(nil, nil)
			},
			room: model.Room{
				ID:        "test-id",
				CreatedAt: t0,
				Name:      "test",
			},
		},

		"Creating a room in a custom database should store the room.": {
			config: mysql.RoomRepositoryConfig{
				Table: "custom-table",
			},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "INSERT INTO custom-table (id, name, created_at) VALUES (?, ?, ?)"
				m.On("ExecContext", mock.Anything, expQuery, "test-id", "test", t0).Once().Return(nil, nil)
			},
			room: model.Room{
				ID:        "test-id",
				CreatedAt: t0,
				Name:      "test",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			mdb := &mysqlmock.DBClient{}
			test.mock(mdb)

			// Execute.
			test.config.DBClient = mdb
			r, err := mysql.NewRoomRepository(test.config)
			require.NoError(err)
			err = r.CreateRoom(context.TODO(), test.room)

			// Check.
			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				mdb.AssertExpectations(t)
			}
		})
	}
}

func TestRoomRepositoryGetRoom(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")
	wantedErr := fmt.Errorf("wanted error")

	tests := map[string]struct {
		config  mysql.RoomRepositoryConfig
		mock    func(*mysqlmock.DBClient)
		id      string
		expRoom *model.Room
		expErr  error
	}{
		"Having an error while retrieving the room should fail.": {
			config: mysql.RoomRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				row := sqlRowErr(wantedErr)
				m.On("QueryRowContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(row)
			},
			id:     "test-id",
			expErr: wantedErr,
		},

		"retrieving a missing room, should fail.": {
			config: mysql.RoomRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				row := sqlRowErr(sql.ErrNoRows)
				m.On("QueryRowContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(row)
			},
			id:     "test-id",
			expErr: internalerrors.ErrMissing,
		},

		"Retrieving a room should get the room.": {
			config: mysql.RoomRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "SELECT room.id, room.name, room.created_at FROM room WHERE id = ?"

				row := sqlmockRowsToStdRow(sqlmock.NewRows([]string{"id", "name", "created_at"}).
					AddRow("test-id", "test", t0))

				m.On("QueryRowContext", mock.Anything, expQuery, "test-id").Once().Return(row)
			},
			id: "test-id",
			expRoom: &model.Room{
				ID:        "test-id",
				CreatedAt: t0,
				Name:      "test",
			},
		},

		"Retrieving a room using custom table should get the room.": {
			config: mysql.RoomRepositoryConfig{
				Table: "custom-table",
			},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "SELECT custom-table.id, custom-table.name, custom-table.created_at FROM custom-table WHERE id = ?"

				row := sqlmockRowsToStdRow(sqlmock.NewRows([]string{"id", "name", "created_at"}).
					AddRow("test-id", "test", t0))

				m.On("QueryRowContext", mock.Anything, expQuery, "test-id").Once().Return(row)
			},
			id: "test-id",
			expRoom: &model.Room{
				ID:        "test-id",
				CreatedAt: t0,
				Name:      "test",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			mdb := &mysqlmock.DBClient{}
			test.mock(mdb)

			// Execute.
			test.config.DBClient = mdb
			r, err := mysql.NewRoomRepository(test.config)
			require.NoError(err)
			gotRoom, err := r.GetRoom(context.TODO(), test.id)

			// Check.
			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				mdb.AssertExpectations(t)
				assert.Equal(test.expRoom, gotRoom)
			}
		})
	}
}

func TestRoomRepositoryRoomExists(t *testing.T) {
	wantedErr := fmt.Errorf("wanted error")

	tests := map[string]struct {
		config    mysql.RoomRepositoryConfig
		mock      func(*mysqlmock.DBClient)
		id        string
		expExists bool
		expErr    error
	}{
		"Having an error while retrieving the room should fail.": {
			config: mysql.RoomRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				row := sqlRowErr(wantedErr)
				m.On("QueryRowContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(row)
			},
			id:     "test-id",
			expErr: wantedErr,
		},

		"Retrieving a existing room should return exists.": {
			config: mysql.RoomRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "SELECT(EXISTS(SELECT * FROM room WHERE id = ?))"
				row := sqlmockRowsToStdRow(sqlmock.NewRows([]string{""}).AddRow(1))

				m.On("QueryRowContext", mock.Anything, expQuery, "test-id").Once().Return(row)
			},
			id:        "test-id",
			expExists: true,
		},

		"Retrieving a non existing room using custom table should return not exists.": {
			config: mysql.RoomRepositoryConfig{
				Table: "custom-table",
			},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "SELECT(EXISTS(SELECT * FROM custom-table WHERE id = ?))"
				row := sqlmockRowsToStdRow(sqlmock.NewRows([]string{""}).AddRow(0))

				m.On("QueryRowContext", mock.Anything, expQuery, "test-id").Once().Return(row)
			},
			id:        "test-id",
			expExists: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			mdb := &mysqlmock.DBClient{}
			test.mock(mdb)

			// Execute.
			test.config.DBClient = mdb
			r, err := mysql.NewRoomRepository(test.config)
			require.NoError(err)
			gotExists, err := r.RoomExists(context.TODO(), test.id)

			// Check.
			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				mdb.AssertExpectations(t)
				assert.Equal(test.expExists, gotExists)
			}
		})
	}
}
