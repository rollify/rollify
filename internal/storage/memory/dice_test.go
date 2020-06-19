package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
	"github.com/rollify/rollify/internal/storage/memory"
)

func TestDiceRollRepositoryCreateDiceRoll(t *testing.T) {
	tests := map[string]struct {
		repo        func() *memory.DiceRollRepository
		diceRoll    model.DiceRoll
		expDiceRoll model.DiceRoll
		expErr      error
	}{

		"Having a dice roll without room ID should return a not valid error.": {
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{
				ID:     "test-id",
				RoomID: "",
				UserID: "user-id",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Having a dice roll without user ID should return a not valid error.": {
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{
				ID:     "test-id",
				RoomID: "room-id",
				UserID: "",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Having a dice roll without ID should return a not valid error.": {
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{
				ID:     "",
				RoomID: "room-id",
				UserID: "user-id",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Creating a dice roll that already exists should return an error.": {
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.DiceRollsByID = map[string]*model.DiceRoll{
					"test-id": {
						ID: "test-id",
					},
				}
				return r
			},
			diceRoll: model.DiceRoll{
				ID:     "test-id",
				RoomID: "room-id",
				UserID: "user-id",
			},
			expErr: internalerrors.ErrAlreadyExists,
		},

		"Creating a dice roll should store the room.": {
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{
				ID:     "test-id",
				RoomID: "room-id",
				UserID: "user-id",
			},
			expDiceRoll: model.DiceRoll{
				ID:     "test-id",
				RoomID: "room-id",
				UserID: "user-id",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			err := r.CreateDiceRoll(context.TODO(), test.diceRoll)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				// Check the dice roll has been created internally.
				// This is very tide to the implementation but we need to access
				// internal structure to see if the implementation actually works.
				gotDiceRoll := r.DiceRollsByID[test.expDiceRoll.ID]
				assert.Equal(test.expDiceRoll, *gotDiceRoll)

				gotDiceRolls := r.DiceRollsByRoom[test.diceRoll.RoomID]
				assert.Contains(gotDiceRolls, &test.expDiceRoll)

				gotDiceRolls = r.DiceRollsByRoomAndUser[test.diceRoll.RoomID+test.diceRoll.UserID]
				assert.Contains(gotDiceRolls, &test.expDiceRoll)
			}
		})
	}
}

func TestDiceRollRepositoryCreateDiceRollSerial(t *testing.T) {
	tests := map[string]struct {
		repo      func() *memory.DiceRollRepository
		diceRolls []model.DiceRoll
	}{

		"creating dice rolls should create unique and incremental serials for each of them.": {
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRolls: []model.DiceRoll{
				{ID: "dr1", UserID: "user1", RoomID: "test"},
				{ID: "dr2", UserID: "user1", RoomID: "test"},
				{ID: "dr3", UserID: "user1", RoomID: "test"},
				{ID: "dr4", UserID: "user1", RoomID: "test"},
				{ID: "dr5", UserID: "user1", RoomID: "test"},
				{ID: "dr6", UserID: "user1", RoomID: "test"},
				{ID: "dr7", UserID: "user1", RoomID: "test"},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			r := test.repo()
			var prevSerial uint
			for i, dr := range test.diceRolls {
				err := r.CreateDiceRoll(context.TODO(), dr)
				require.NoError(err)

				// Don't check first one.
				if i == 0 {
					prevSerial = dr.Serial
					continue
				}

				// Check our dr serial is greater than the previous one.
				storedDiceRoll := r.DiceRollsByID[dr.ID]
				assert.Greater(storedDiceRoll.Serial, prevSerial)
				prevSerial = storedDiceRoll.Serial
			}

		})
	}
}

func TestDiceRollRepositoryListDiceRoll(t *testing.T) {
	tests := map[string]struct {
		pageOpts     model.PaginationOpts
		filterOpts   storage.ListDiceRollsOpts
		repo         func() *memory.DiceRollRepository
		expDiceRolls *storage.DiceRollList
		expErr       error
	}{
		"Listing dice rolls without room should fail.": {
			pageOpts: model.PaginationOpts{Size: 9999999999},
			filterOpts: storage.ListDiceRollsOpts{
				RoomID: "",
				UserID: "",
			},
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Listing all users dice rolls in a room should return all of the room dice rolls.": {
			pageOpts: model.PaginationOpts{Size: 9999999999},
			filterOpts: storage.ListDiceRollsOpts{
				RoomID: "room-1",
				UserID: "",
			},
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.DiceRollsByRoom = map[string][]*model.DiceRoll{
					"room-0": {{ID: "test00", Serial: 0}},
					"room-1": {{ID: "test10", Serial: 1}, {ID: "test11", Serial: 2}},
					"room-2": {{ID: "test20", Serial: 3}},
				}
				return r
			},
			expDiceRolls: &storage.DiceRollList{
				Items: []model.DiceRoll{
					{ID: "test11", Serial: 2},
					{ID: "test10", Serial: 1},
				},
				Cursors: model.PaginationCursors{
					FirstCursor: "eyJzZXJpYWwiOjJ9",
					LastCursor:  "eyJzZXJpYWwiOjF9",
					HasPrevious: false,
					HasNext:     false,
				},
			},
		},

		"Listing single user dice rolls in a room should return only the onest of the user in the room.": {
			pageOpts: model.PaginationOpts{Size: 9999999999},
			filterOpts: storage.ListDiceRollsOpts{
				RoomID: "room-2",
				UserID: "user-1",
			},
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.DiceRollsByRoomAndUser = map[string][]*model.DiceRoll{
					"room-0user-1": {{ID: "test00", Serial: 0}},
					"room-1user-2": {{ID: "test10", Serial: 1}, {ID: "test11", Serial: 2}},
					"room-2user-1": {{ID: "test20", Serial: 3}},
				}
				return r
			},
			expDiceRolls: &storage.DiceRollList{
				Items: []model.DiceRoll{
					{ID: "test20", Serial: 3},
				},
				Cursors: model.PaginationCursors{
					FirstCursor: "eyJzZXJpYWwiOjN9",
					LastCursor:  "eyJzZXJpYWwiOjN9",
					HasPrevious: false,
					HasNext:     false,
				},
			},
		},

		"Listing dice rolls with a cursor and limit should return them by default in desc order.": {
			pageOpts: model.PaginationOpts{
				Cursor: "eyJzZXJpYWwiOjN9",
				Size:   2,
			},
			filterOpts: storage.ListDiceRollsOpts{RoomID: "room-0"},
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.DiceRollsByRoom = map[string][]*model.DiceRoll{
					"room-0": {
						{ID: "test00", Serial: 0},
						{ID: "test01", Serial: 1},
						{ID: "test02", Serial: 2},
						{ID: "test03", Serial: 3},
						{ID: "test04", Serial: 4},
						{ID: "test05", Serial: 5},
						{ID: "test06", Serial: 6},
					},
				}
				return r
			},
			expDiceRolls: &storage.DiceRollList{
				Items: []model.DiceRoll{
					{ID: "test02", Serial: 2},
					{ID: "test01", Serial: 1},
				},
				Cursors: model.PaginationCursors{
					FirstCursor: "eyJzZXJpYWwiOjJ9",
					LastCursor:  "eyJzZXJpYWwiOjF9",
					HasPrevious: true,
					HasNext:     true,
				},
			},
		},

		"Listing dice rolls in asc order, should return them in asc order.": {
			pageOpts: model.PaginationOpts{
				Cursor: "eyJzZXJpYWwiOjN9",
				Size:   2,
				Order:  model.PaginationOrderAsc,
			},
			filterOpts: storage.ListDiceRollsOpts{RoomID: "room-0"},
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.DiceRollsByRoom = map[string][]*model.DiceRoll{
					"room-0": {
						{ID: "test00", Serial: 0},
						{ID: "test01", Serial: 1},
						{ID: "test02", Serial: 2},
						{ID: "test03", Serial: 3},
						{ID: "test04", Serial: 4},
						{ID: "test05", Serial: 5},
						{ID: "test06", Serial: 6},
					},
				}
				return r
			},
			expDiceRolls: &storage.DiceRollList{
				Items: []model.DiceRoll{
					{ID: "test04", Serial: 4},
					{ID: "test05", Serial: 5},
				},
				Cursors: model.PaginationCursors{
					FirstCursor: "eyJzZXJpYWwiOjR9",
					LastCursor:  "eyJzZXJpYWwiOjV9",
					HasPrevious: true,
					HasNext:     true,
				},
			},
		},

		"Listing dice rolls that needs to return all of them, should return that doesn't have more.": {
			pageOpts: model.PaginationOpts{
				Cursor: "eyJzZXJpYWwiOjN9",
				Size:   9999999999,
				Order:  model.PaginationOrderAsc,
			},
			filterOpts: storage.ListDiceRollsOpts{RoomID: "room-0"},
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.DiceRollsByRoom = map[string][]*model.DiceRoll{
					"room-0": {
						{ID: "test00", Serial: 0},
						{ID: "test01", Serial: 1},
						{ID: "test02", Serial: 2},
						{ID: "test03", Serial: 3},
						{ID: "test04", Serial: 4},
						{ID: "test05", Serial: 5},
						{ID: "test06", Serial: 6},
					},
				}
				return r
			},
			expDiceRolls: &storage.DiceRollList{
				Items: []model.DiceRoll{
					{ID: "test04", Serial: 4},
					{ID: "test05", Serial: 5},
					{ID: "test06", Serial: 6},
				},
				Cursors: model.PaginationCursors{
					FirstCursor: "eyJzZXJpYWwiOjR9",
					LastCursor:  "eyJzZXJpYWwiOjZ9",
					HasPrevious: true,
					HasNext:     false,
				},
			},
		},

		"Listing all dice roolls without cursors in asc order.": {
			pageOpts: model.PaginationOpts{
				Size:  100,
				Order: model.PaginationOrderAsc,
			},
			filterOpts: storage.ListDiceRollsOpts{RoomID: "room-0"},
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.DiceRollsByRoom = map[string][]*model.DiceRoll{
					"room-0": {
						{ID: "test00", Serial: 0},
						{ID: "test01", Serial: 1},
						{ID: "test02", Serial: 2},
						{ID: "test03", Serial: 3},
						{ID: "test04", Serial: 4},
						{ID: "test05", Serial: 5},
						{ID: "test06", Serial: 6},
					},
				}
				return r
			},
			expDiceRolls: &storage.DiceRollList{
				Items: []model.DiceRoll{
					{ID: "test00", Serial: 0},
					{ID: "test01", Serial: 1},
					{ID: "test02", Serial: 2},
					{ID: "test03", Serial: 3},
					{ID: "test04", Serial: 4},
					{ID: "test05", Serial: 5},
					{ID: "test06", Serial: 6},
				},
				Cursors: model.PaginationCursors{
					FirstCursor: "eyJzZXJpYWwiOjB9",
					LastCursor:  "eyJzZXJpYWwiOjZ9",
					HasPrevious: false,
					HasNext:     false,
				},
			},
		},

		"Listing all dice roolls without cursors in desc order.": {
			pageOpts: model.PaginationOpts{
				Size:  100,
				Order: model.PaginationOrderDesc,
			},
			filterOpts: storage.ListDiceRollsOpts{RoomID: "room-0"},
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.DiceRollsByRoom = map[string][]*model.DiceRoll{
					"room-0": {
						{ID: "test00", Serial: 0},
						{ID: "test01", Serial: 1},
						{ID: "test02", Serial: 2},
						{ID: "test03", Serial: 3},
						{ID: "test04", Serial: 4},
						{ID: "test05", Serial: 5},
						{ID: "test06", Serial: 6},
					},
				}
				return r
			},
			expDiceRolls: &storage.DiceRollList{
				Items: []model.DiceRoll{
					{ID: "test06", Serial: 6},
					{ID: "test05", Serial: 5},
					{ID: "test04", Serial: 4},
					{ID: "test03", Serial: 3},
					{ID: "test02", Serial: 2},
					{ID: "test01", Serial: 1},
					{ID: "test00", Serial: 0},
				},
				Cursors: model.PaginationCursors{
					FirstCursor: "eyJzZXJpYWwiOjZ9",
					LastCursor:  "eyJzZXJpYWwiOjB9",
					HasPrevious: false,
					HasNext:     false,
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			gotDiceRolls, err := r.ListDiceRolls(context.TODO(), test.pageOpts, test.filterOpts)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				assert.Equal(test.expDiceRolls, gotDiceRolls)
			}
		})
	}
}
