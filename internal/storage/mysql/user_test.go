package mysql_test

import (
	"context"
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
	"github.com/rollify/rollify/internal/storage"
	"github.com/rollify/rollify/internal/storage/mysql"
	"github.com/rollify/rollify/internal/storage/mysql/mysqlmock"
)

func TestUserRepositoryCreateUser(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")
	wantedErr := fmt.Errorf("wanted error")

	tests := map[string]struct {
		config mysql.UserRepositoryConfig
		mock   func(*mysqlmock.DBClient)
		user   model.User
		expErr error
	}{
		"Having an error while storing the user, should error.": {
			config: mysql.UserRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				m.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, wantedErr)
			},
			user: model.User{
				ID:        "test-id",
				RoomID:    "room-id",
				CreatedAt: t0,
				Name:      "test",
			},
			expErr: wantedErr,
		},

		"Creating the same user when already exists, should error.": {
			config: mysql.UserRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				err := &drivermysql.MySQLError{Number: 1062}
				m.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, err)
			},
			user: model.User{
				ID:        "test-id",
				RoomID:    "room-id",
				CreatedAt: t0,
				Name:      "test",
			},
			expErr: internalerrors.ErrAlreadyExists,
		},

		"Creating a user should store the user.": {
			config: mysql.UserRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "INSERT INTO user (id, name, room_id, created_at) VALUES (?, ?, ?, ?)"
				m.On("ExecContext", mock.Anything, expQuery, "test-id", "test", "room-id", t0).Once().Return(nil, nil)
			},
			user: model.User{
				ID:        "test-id",
				RoomID:    "room-id",
				CreatedAt: t0,
				Name:      "test",
			},
		},

		"Creating a user in a custom database should store the user.": {
			config: mysql.UserRepositoryConfig{
				Table: "custom-table",
			},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "INSERT INTO custom-table (id, name, room_id, created_at) VALUES (?, ?, ?, ?)"
				m.On("ExecContext", mock.Anything, expQuery, "test-id", "test", "room-id", t0).Once().Return(nil, nil)
			},
			user: model.User{
				ID:        "test-id",
				RoomID:    "room-id",
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
			r, err := mysql.NewUserRepository(test.config)
			require.NoError(err)
			err = r.CreateUser(context.TODO(), test.user)

			// Check.
			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				mdb.AssertExpectations(t)
			}
		})
	}
}

func TestUserRepositoryListUsers(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")
	wantedErr := fmt.Errorf("wanted error")

	tests := map[string]struct {
		config      mysql.UserRepositoryConfig
		mock        func(*mysqlmock.DBClient)
		roomID      string
		expUserList *storage.UserList
		expErr      error
	}{
		"Having an error while retrieving listing the users, should fail.": {
			config: mysql.UserRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				m.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, wantedErr)
			},
			roomID: "test-id",
			expErr: wantedErr,
		},

		"Retrieving the users with rows error should fail.": {
			config: mysql.UserRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				rows := sqlmockRowsToStdRows(sqlmock.NewRows([]string{"id", "name", "room_id", "created_at"}).
					AddRow("test0-id", "test0", "room-id", t0).
					RowError(0, wantedErr))

				m.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
			},
			roomID: "room-id",
			expErr: wantedErr,
		},

		"Retrieving the users from a room should get the users.": {
			config: mysql.UserRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "SELECT user.id, user.name, user.room_id, user.created_at FROM user WHERE room_id = ?"

				rows := sqlmockRowsToStdRows(sqlmock.NewRows([]string{"id", "name", "room_id", "created_at"}).
					AddRow("test0-id", "test0", "room-id", t0).
					AddRow("test1-id", "test1", "room-id", t0).
					AddRow("test2-id", "test2", "room-id", t0).
					AddRow("test3-id", "", "room-id", t0).
					AddRow("test4-id", "test4", "room-id", t0))

				m.On("QueryContext", mock.Anything, expQuery, "room-id").Once().Return(rows, nil)
			},
			roomID: "room-id",
			expUserList: &storage.UserList{
				Items: []model.User{
					{ID: "test0-id", Name: "test0", RoomID: "room-id", CreatedAt: t0},
					{ID: "test1-id", Name: "test1", RoomID: "room-id", CreatedAt: t0},
					{ID: "test2-id", Name: "test2", RoomID: "room-id", CreatedAt: t0},
					{ID: "test3-id", Name: "", RoomID: "room-id", CreatedAt: t0},
					{ID: "test4-id", Name: "test4", RoomID: "room-id", CreatedAt: t0},
				},
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
			r, err := mysql.NewUserRepository(test.config)
			require.NoError(err)
			gotUserList, err := r.ListRoomUsers(context.TODO(), test.roomID)

			// Check.
			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				mdb.AssertExpectations(t)
				assert.Equal(test.expUserList, gotUserList)
			}
		})
	}
}

func TestUserRepositoryUserExists(t *testing.T) {
	wantedErr := fmt.Errorf("wanted error")

	tests := map[string]struct {
		config    mysql.UserRepositoryConfig
		mock      func(*mysqlmock.DBClient)
		id        string
		expExists bool
		expErr    error
	}{
		"Having an error while retrieving the user should fail.": {
			config: mysql.UserRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				row := sqlRowErr(wantedErr)
				m.On("QueryRowContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(row)
			},
			id:     "test-id",
			expErr: wantedErr,
		},

		"Retrieving a existing user using should return exists.": {
			config: mysql.UserRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "SELECT(EXISTS(SELECT * FROM user WHERE id = ?))"
				row := sqlmockRowsToStdRow(sqlmock.NewRows([]string{""}).AddRow(1))

				m.On("QueryRowContext", mock.Anything, expQuery, "test-id").Once().Return(row)
			},
			id:        "test-id",
			expExists: true,
		},

		"Retrieving a non existing user using custom table should return not exists.": {
			config: mysql.UserRepositoryConfig{
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
			r, err := mysql.NewUserRepository(test.config)
			require.NoError(err)
			gotExists, err := r.UserExists(context.TODO(), test.id)

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

func TestUserRepositoryUserExistsByNameInsensitive(t *testing.T) {
	wantedErr := fmt.Errorf("wanted error")

	tests := map[string]struct {
		config    mysql.UserRepositoryConfig
		mock      func(*mysqlmock.DBClient)
		roomID    string
		username  string
		expExists bool
		expErr    error
	}{
		"Having an error while retrieving the user should fail.": {
			config: mysql.UserRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				row := sqlRowErr(wantedErr)
				m.On("QueryRowContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(row)
			},
			roomID:   "room-id",
			username: "test-user",
			expErr:   wantedErr,
		},

		"Retrieving a existing user should return exists.": {
			config: mysql.UserRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "SELECT(EXISTS(SELECT * FROM user WHERE room_id = ? AND name = ?))"
				row := sqlmockRowsToStdRow(sqlmock.NewRows([]string{""}).AddRow(1))

				m.On("QueryRowContext", mock.Anything, expQuery, "room-id", "test-user").Once().Return(row)
			},
			roomID:    "room-id",
			username:  "test-user",
			expExists: true,
		},

		"Retrieving a non existing user using custom table should return not exists.": {
			config: mysql.UserRepositoryConfig{
				Table: "custom-table",
			},
			mock: func(m *mysqlmock.DBClient) {
				expQuery := "SELECT(EXISTS(SELECT * FROM custom-table WHERE room_id = ? AND name = ?))"
				row := sqlmockRowsToStdRow(sqlmock.NewRows([]string{""}).AddRow(0))

				m.On("QueryRowContext", mock.Anything, expQuery, "room-id", "test-user").Once().Return(row)
			},
			roomID:    "room-id",
			username:  "test-user",
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
			r, err := mysql.NewUserRepository(test.config)
			require.NoError(err)
			gotExists, err := r.UserExistsByNameInsensitive(context.TODO(), test.roomID, test.username)

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
