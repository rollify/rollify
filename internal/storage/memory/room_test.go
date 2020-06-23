package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage/memory"
)

func TestRoomRepositoryCreateRoom(t *testing.T) {
	tests := map[string]struct {
		repo    func() *memory.RoomRepository
		room    model.Room
		expRoom model.Room
		expErr  error
	}{
		"Having a room without ID should return a not valid error.": {
			repo: func() *memory.RoomRepository {
				return memory.NewRoomRepository()
			},
			room: model.Room{
				ID:   "",
				Name: "test",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Creating a room that already exists should return an error.": {
			repo: func() *memory.RoomRepository {
				r := memory.NewRoomRepository()
				r.RoomsByID = map[string]*model.Room{
					"test-id": {
						ID:   "test-id",
						Name: "test",
					},
				}
				return r
			},
			room: model.Room{
				ID:   "test-id",
				Name: "test",
			},
			expErr: internalerrors.ErrAlreadyExists,
		},

		"Creating a room should store the room.": {
			repo: func() *memory.RoomRepository {
				return memory.NewRoomRepository()
			},
			room: model.Room{
				ID:   "test-id",
				Name: "test",
			},
			expRoom: model.Room{
				ID:   "test-id",
				Name: "test",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			err := r.CreateRoom(context.TODO(), test.room)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				// Check the room has been created internally.
				gotRoom := r.RoomsByID[test.expRoom.ID]
				assert.Equal(test.expRoom, *gotRoom)
			}
		})
	}
}

func TestRoomRepositoryGetRoom(t *testing.T) {
	tests := map[string]struct {
		repo    func() *memory.RoomRepository
		roomID  string
		expRoom *model.Room
		expErr  error
	}{
		"Having a room ID that does not exists in the repository, it should return an error.": {
			repo: func() *memory.RoomRepository {
				return memory.NewRoomRepository()
			},
			roomID: "test-id",
			expErr: internalerrors.ErrMissing,
		},

		"Having a room ID that exists in the repository, it should return exists.": {
			repo: func() *memory.RoomRepository {
				r := memory.NewRoomRepository()
				r.RoomsByID = map[string]*model.Room{
					"test-id": {ID: "test-id", Name: "test"},
				}
				return r
			},
			roomID: "test-id",
			expRoom: &model.Room{
				ID:   "test-id",
				Name: "test",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			gotRoom, err := r.GetRoom(context.TODO(), test.roomID)

			if test.expErr != nil {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				assert.Equal(test.expRoom, gotRoom)
			}
		})
	}
}

func TestRoomRepositoryRoomExists(t *testing.T) {
	tests := map[string]struct {
		repo      func() *memory.RoomRepository
		roomID    string
		expExists bool
		expErr    bool
	}{
		"Having a room ID that does not exists in the repository, it should return no exists.": {
			repo: func() *memory.RoomRepository {
				return memory.NewRoomRepository()
			},
			roomID:    "test-id",
			expExists: false,
		},

		"Having a room ID that exists in the repository, it should return exists.": {
			repo: func() *memory.RoomRepository {
				r := memory.NewRoomRepository()
				r.RoomsByID = map[string]*model.Room{
					"test-id": {ID: "test-id", Name: "test"},
				}
				return r
			},
			roomID:    "test-id",
			expExists: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			gotExists, err := r.RoomExists(context.TODO(), test.roomID)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expExists, gotExists)
			}
		})
	}
}
