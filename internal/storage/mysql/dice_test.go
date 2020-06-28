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

func TestDiceRollRepositoryCreateDiceRoll(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")
	wantedErr := fmt.Errorf("wanted error")

	tests := map[string]struct {
		config   mysql.DiceRollRepositoryConfig
		mock     func(*mysqlmock.DBClient)
		diceRoll model.DiceRoll
		expErr   error
	}{
		"Having an error while storing the dice roll, should error.": {
			config: mysql.DiceRollRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				m.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, wantedErr)
			},
			diceRoll: model.DiceRoll{
				ID:        "dice-roll-id",
				RoomID:    "room-id",
				UserID:    "user-id",
				CreatedAt: t0,
				Dice:      []model.DieRoll{},
			},
			expErr: wantedErr,
		},

		"Creating the same dice roll when already exists, should error.": {
			config: mysql.DiceRollRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				err := &drivermysql.MySQLError{Number: 1062}
				m.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, err)
			},
			diceRoll: model.DiceRoll{
				ID:        "dice-roll-id",
				RoomID:    "room-id",
				UserID:    "user-id",
				CreatedAt: t0,
				Dice:      []model.DieRoll{},
			},
			expErr: internalerrors.ErrAlreadyExists,
		},

		"Having an error while storing the die rolls, should error.": {
			config: mysql.DiceRollRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				m.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				m.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, wantedErr)
			},
			diceRoll: model.DiceRoll{
				ID:        "dice-roll-id",
				RoomID:    "room-id",
				UserID:    "user-id",
				CreatedAt: t0,
				Dice:      []model.DieRoll{{ID: "dr1", Type: model.DieTypeD6, Side: 5}},
			},
			expErr: wantedErr,
		},

		"Creating the same die rolls that already exist, should error.": {
			config: mysql.DiceRollRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				err := &drivermysql.MySQLError{Number: 1062}
				m.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				m.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, err)
			},
			diceRoll: model.DiceRoll{
				ID:        "dice-roll-id",
				RoomID:    "room-id",
				UserID:    "user-id",
				CreatedAt: t0,
				Dice:      []model.DieRoll{{ID: "dr1", Type: model.DieTypeD6, Side: 5}},
			},
			expErr: internalerrors.ErrAlreadyExists,
		},

		"Creating the a dice roll should store the dice roll and die rolls.": {
			config: mysql.DiceRollRepositoryConfig{},
			mock: func(m *mysqlmock.DBClient) {
				// Expected dice roll.
				expQuery := "INSERT INTO dice_roll (id, created_at, room_id, user_id) VALUES (?, ?, ?, ?)"
				m.On("ExecContext", mock.Anything, expQuery, "dice-roll-id", t0, "room-id", "user-id").Once().Return(nil, nil)

				// Expected die rolls.
				expQuery = "INSERT INTO die_roll (id, dice_roll_id, die_type_id, side) VALUES (?, ?, ?, ?), (?, ?, ?, ?), (?, ?, ?, ?)"
				m.On("ExecContext", mock.Anything, expQuery,
					"dr1", "dice-roll-id", "d6", uint(5),
					"dr2", "dice-roll-id", "d10", uint(7),
					"dr3", "dice-roll-id", "d20", uint(15),
				).Once().Return(nil, nil)
			},
			diceRoll: model.DiceRoll{
				ID:        "dice-roll-id",
				RoomID:    "room-id",
				UserID:    "user-id",
				CreatedAt: t0,
				Dice: []model.DieRoll{
					{ID: "dr1", Type: model.DieTypeD6, Side: 5},
					{ID: "dr2", Type: model.DieTypeD10, Side: 7},
					{ID: "dr3", Type: model.DieTypeD20, Side: 15},
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
			r, err := mysql.NewDiceRollRepository(test.config)
			require.NoError(err)
			err = r.CreateDiceRoll(context.TODO(), test.diceRoll)

			// Check.
			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				mdb.AssertExpectations(t)
			}
		})
	}
}

func TestDiceRollRepositoryListUsers(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")
	wantedErr := fmt.Errorf("wanted error")

	tests := map[string]struct {
		config          mysql.DiceRollRepositoryConfig
		pageOpts        model.PaginationOpts
		filterOpts      storage.ListDiceRollsOpts
		mock            func(*mysqlmock.DBClient)
		roomID          string
		expDiceRollList *storage.DiceRollList
		expErr          error
	}{
		"Having a bad encoded cursor should error.": {
			pageOpts: model.PaginationOpts{
				Cursor: "wrong",
			},
			filterOpts: storage.ListDiceRollsOpts{
				RoomID: "room-1",
			},
			mock:   func(m *mysqlmock.DBClient) {},
			expErr: internalerrors.ErrNotValid,
		},

		"Having a bad formated cursor should error.": {
			pageOpts: model.PaginationOpts{
				Cursor: "ew==",
			},
			filterOpts: storage.ListDiceRollsOpts{
				RoomID: "room-1",
			},
			mock:   func(m *mysqlmock.DBClient) {},
			expErr: internalerrors.ErrNotValid,
		},

		"Having an error while retrieving the dice rolls should fail.": {
			mock: func(m *mysqlmock.DBClient) {
				m.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, wantedErr)
			},
			roomID: "test-id",
			expErr: wantedErr,
		},

		"Listing all room dice rolls in a room should return all of the room dice rolls correctly mapped.": {
			pageOpts: model.PaginationOpts{},
			filterOpts: storage.ListDiceRollsOpts{
				RoomID: "room-1",
				UserID: "",
			},
			mock: func(m *mysqlmock.DBClient) {
				rows := sqlmockRowsToStdRows(sqlmock.NewRows([]string{"drs.id", "drs.created_at", "drs.room_id", "drs.user_id", "drs.serial", "dr.id", "dr.die_type_id", "dr.side"}).
					AddRow("dr2", t0, "room-1", "user-2", 3, "dr20", "d20", 11).
					AddRow("dr2", t0, "room-1", "user-2", 3, "dr21", "d20", 17).
					AddRow("dr1", t0, "room-1", "user-1", 2, "dr10", "d10", 8).
					AddRow("dr0", t0, "room-1", "user-1", 1, "dr00", "d6", 0).
					AddRow("dr0", t0, "room-1", "user-1", 1, "dr01", "d6", 4))
				// Expected dice roll.
				expQuery := "SELECT drs.id, drs.created_at, drs.room_id, drs.user_id, drs.serial, dr.id, dr.die_type_id, dr.side FROM die_roll dr JOIN (SELECT id, created_at, room_id, user_id, serial FROM dice_roll WHERE room_id = ? ORDER BY serial DESC) AS drs ON dr.dice_roll_id = drs.id ORDER BY serial DESC"
				m.On("QueryContext", mock.Anything, expQuery, "room-1").Once().Return(rows, nil)
			},
			expDiceRollList: &storage.DiceRollList{
				Items: []model.DiceRoll{
					{ID: "dr2", RoomID: "room-1", CreatedAt: t0, Serial: 3, UserID: "user-2",
						Dice: []model.DieRoll{
							{ID: "dr20", Type: model.DieTypeD20, Side: 11},
							{ID: "dr21", Type: model.DieTypeD20, Side: 17},
						},
					},
					{ID: "dr1", RoomID: "room-1", CreatedAt: t0, Serial: 2, UserID: "user-1",
						Dice: []model.DieRoll{
							{ID: "dr10", Type: model.DieTypeD10, Side: 8},
						},
					},
					{ID: "dr0", RoomID: "room-1", CreatedAt: t0, Serial: 1, UserID: "user-1",
						Dice: []model.DieRoll{
							{ID: "dr00", Type: model.DieTypeD6, Side: 0},
							{ID: "dr01", Type: model.DieTypeD6, Side: 4},
						},
					},
				},
				Cursors: model.PaginationCursors{
					FirstCursor: "eyJzZXJpYWwiOjN9",
					LastCursor:  "eyJzZXJpYWwiOjF9",
					HasPrevious: false,
					HasNext:     true,
				},
			},
		},

		"Listing specific user dice rolls in a room should filter by user.": {
			pageOpts: model.PaginationOpts{},
			filterOpts: storage.ListDiceRollsOpts{
				RoomID: "room-1",
				UserID: "user-1",
			},
			mock: func(m *mysqlmock.DBClient) {
				// Expected dice roll.
				rows := sqlmockRowsToStdRows(sqlmock.NewRows([]string{"drs.id", "drs.created_at", "drs.room_id", "drs.user_id", "drs.serial", "dr.id", "dr.die_type_id", "dr.side"}))
				expQuery := "SELECT drs.id, drs.created_at, drs.room_id, drs.user_id, drs.serial, dr.id, dr.die_type_id, dr.side FROM die_roll dr JOIN (SELECT id, created_at, room_id, user_id, serial FROM dice_roll WHERE room_id = ? AND drs.user_id = ? ORDER BY serial DESC) AS drs ON dr.dice_roll_id = drs.id ORDER BY serial DESC"
				m.On("QueryContext", mock.Anything, expQuery, "room-1", "user-1").Once().Return(rows, nil)
			},
			expDiceRollList: &storage.DiceRollList{
				Items: []model.DiceRoll{},
				Cursors: model.PaginationCursors{
					FirstCursor: "",
					LastCursor:  "",
					HasPrevious: false,
					HasNext:     true,
				},
			},
		},

		"Listing dice rolls in asc mode a room should filter order them by asc.": {
			pageOpts: model.PaginationOpts{
				Order: model.PaginationOrderAsc,
			},
			filterOpts: storage.ListDiceRollsOpts{
				RoomID: "room-1",
			},
			mock: func(m *mysqlmock.DBClient) {
				// Expected dice roll.
				rows := sqlmockRowsToStdRows(sqlmock.NewRows([]string{"drs.id", "drs.created_at", "drs.room_id", "drs.user_id", "drs.serial", "dr.id", "dr.die_type_id", "dr.side"}))
				expQuery := "SELECT drs.id, drs.created_at, drs.room_id, drs.user_id, drs.serial, dr.id, dr.die_type_id, dr.side FROM die_roll dr JOIN (SELECT id, created_at, room_id, user_id, serial FROM dice_roll WHERE room_id = ? ORDER BY serial ASC) AS drs ON dr.dice_roll_id = drs.id ORDER BY serial ASC"
				m.On("QueryContext", mock.Anything, expQuery, "room-1").Once().Return(rows, nil)
			},
			expDiceRollList: &storage.DiceRollList{
				Items: []model.DiceRoll{},
				Cursors: model.PaginationCursors{
					FirstCursor: "",
					LastCursor:  "",
					HasPrevious: false,
					HasNext:     true,
				},
			},
		},

		"Listing dice rolls in with size mode a room should limit the results.": {
			pageOpts: model.PaginationOpts{
				Size: 42,
			},
			filterOpts: storage.ListDiceRollsOpts{
				RoomID: "room-1",
			},
			mock: func(m *mysqlmock.DBClient) {
				// Expected dice roll.
				rows := sqlmockRowsToStdRows(sqlmock.NewRows([]string{"drs.id", "drs.created_at", "drs.room_id", "drs.user_id", "drs.serial", "dr.id", "dr.die_type_id", "dr.side"}))
				expQuery := "SELECT drs.id, drs.created_at, drs.room_id, drs.user_id, drs.serial, dr.id, dr.die_type_id, dr.side FROM die_roll dr JOIN (SELECT id, created_at, room_id, user_id, serial FROM dice_roll WHERE room_id = ? ORDER BY serial DESC LIMIT 42) AS drs ON dr.dice_roll_id = drs.id ORDER BY serial DESC"
				m.On("QueryContext", mock.Anything, expQuery, "room-1").Once().Return(rows, nil)
			},
			expDiceRollList: &storage.DiceRollList{
				Items: []model.DiceRoll{},
				Cursors: model.PaginationCursors{
					FirstCursor: "",
					LastCursor:  "",
					HasPrevious: false,
					HasNext:     false,
				},
			},
		},

		"Listing dice rolls with cursor in desc mode should list from the cursor order by desc.": {
			pageOpts: model.PaginationOpts{
				Cursor: "eyJzZXJpYWwiOjN9",
				Order:  model.PaginationOrderDesc,
			},
			filterOpts: storage.ListDiceRollsOpts{
				RoomID: "room-1",
			},
			mock: func(m *mysqlmock.DBClient) {
				// Expected dice roll.
				rows := sqlmockRowsToStdRows(sqlmock.NewRows([]string{"drs.id", "drs.created_at", "drs.room_id", "drs.user_id", "drs.serial", "dr.id", "dr.die_type_id", "dr.side"}))
				expQuery := "SELECT drs.id, drs.created_at, drs.room_id, drs.user_id, drs.serial, dr.id, dr.die_type_id, dr.side FROM die_roll dr JOIN (SELECT id, created_at, room_id, user_id, serial FROM dice_roll WHERE room_id = ? AND serial < ? ORDER BY serial DESC) AS drs ON dr.dice_roll_id = drs.id ORDER BY serial DESC"
				m.On("QueryContext", mock.Anything, expQuery, "room-1", 3).Once().Return(rows, nil)
			},
			expDiceRollList: &storage.DiceRollList{
				Items: []model.DiceRoll{},
				Cursors: model.PaginationCursors{
					FirstCursor: "",
					LastCursor:  "",
					HasPrevious: true,
					HasNext:     true,
				},
			},
		},

		"Listing dice rolls with cursor in asc mode should list from the cursor order by asc.": {
			pageOpts: model.PaginationOpts{
				Cursor: "eyJzZXJpYWwiOjN9",
				Order:  model.PaginationOrderAsc,
			},
			filterOpts: storage.ListDiceRollsOpts{
				RoomID: "room-1",
			},
			mock: func(m *mysqlmock.DBClient) {
				// Expected dice roll.
				rows := sqlmockRowsToStdRows(sqlmock.NewRows([]string{"drs.id", "drs.created_at", "drs.room_id", "drs.user_id", "drs.serial", "dr.id", "dr.die_type_id", "dr.side"}))
				expQuery := "SELECT drs.id, drs.created_at, drs.room_id, drs.user_id, drs.serial, dr.id, dr.die_type_id, dr.side FROM die_roll dr JOIN (SELECT id, created_at, room_id, user_id, serial FROM dice_roll WHERE room_id = ? AND serial > ? ORDER BY serial ASC) AS drs ON dr.dice_roll_id = drs.id ORDER BY serial ASC"
				m.On("QueryContext", mock.Anything, expQuery, "room-1", 3).Once().Return(rows, nil)
			},
			expDiceRollList: &storage.DiceRollList{
				Items: []model.DiceRoll{},
				Cursors: model.PaginationCursors{
					FirstCursor: "",
					LastCursor:  "",
					HasPrevious: true,
					HasNext:     true,
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
			r, err := mysql.NewDiceRollRepository(test.config)
			require.NoError(err)
			gotDiceRollList, err := r.ListDiceRolls(context.TODO(), test.pageOpts, test.filterOpts)

			// Check.
			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				mdb.AssertExpectations(t)
				assert.Equal(test.expDiceRollList, gotDiceRollList)
			}
		})
	}
}
