package mysql_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

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
	//t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")
	//wantedErr := fmt.Errorf("wanted error")

	tests := map[string]struct {
		config          mysql.DiceRollRepositoryConfig
		pageOpts        model.PaginationOpts
		filterOpts      storage.ListDiceRollsOpts
		mock            func(*mysqlmock.DBClient)
		roomID          string
		expDiceRollList *storage.DiceRollList
		expErr          error
	}{}

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
