package memory_test

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
	"github.com/rollify/rollify/internal/storage/memory"
)

func TestUserRepositoryCreateUser(t *testing.T) {
	tests := map[string]struct {
		repo    func() *memory.UserRepository
		user    model.User
		expUser model.User
		expErr  error
	}{
		"Having a user without ID should return a not valid error.": {
			repo: func() *memory.UserRepository {
				return memory.NewUserRepository()
			},
			user: model.User{
				ID:     "",
				RoomID: "room-id",
				Name:   "test",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Having a user without room ID should return a not valid error.": {
			repo: func() *memory.UserRepository {
				return memory.NewUserRepository()
			},
			user: model.User{
				ID:     "user-id",
				RoomID: "",
				Name:   "test",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Having a user without name should return a not valid error.": {
			repo: func() *memory.UserRepository {
				return memory.NewUserRepository()
			},
			user: model.User{
				ID:     "user-id",
				RoomID: "room-id",
				Name:   "",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Having an already stored user should be return an error.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByRoom = map[string]map[string]*model.User{
					"room-id": {
						"user-id": &model.User{
							ID:     "user-id",
							RoomID: "room-id",
							Name:   "tst",
						},
					},
				}
				return r
			},
			user: model.User{
				ID:     "user-id",
				RoomID: "room-id",
				Name:   "tst",
			},
			expErr: internalerrors.ErrAlreadyExists,
		},

		"Having a user should be stored.": {
			repo: func() *memory.UserRepository {
				return memory.NewUserRepository()
			},
			user: model.User{
				ID:     "user-id",
				RoomID: "room-id",
				Name:   "tst",
			},
			expUser: model.User{
				ID:     "user-id",
				RoomID: "room-id",
				Name:   "tst",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			r := test.repo()
			err := r.CreateUser(context.TODO(), test.user)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				// Check the user has been created internally.
				usersByRoom := r.UsersByRoom[test.expUser.RoomID]
				require.NotNil(usersByRoom)
				gotUser := usersByRoom[test.expUser.ID]
				require.NotNil(gotUser)
				assert.Equal(test.expUser, *gotUser)

				gotUser = r.UsersByID[test.expUser.ID]
				require.NotNil(gotUser)
				assert.Equal(test.expUser, *gotUser)
			}
		})
	}
}

func TestUserRepositoryListRoomUsers(t *testing.T) {
	tests := map[string]struct {
		repo    func() *memory.UserRepository
		roomID  string
		expList *storage.UserList
		expErr  bool
	}{
		"Using an empty room ID should return an error.": {
			repo: func() *memory.UserRepository {
				return memory.NewUserRepository()
			},
			roomID: "",
			expErr: true,
		},

		"Using a room ID should return that room users.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByRoom = map[string]map[string]*model.User{
					"room1-id": {
						"user1-id": &model.User{
							ID:     "user1-id",
							RoomID: "room1-id",
							Name:   "test1",
						},
					},

					"room2-id": {
						"user2-id": &model.User{
							ID:     "user2-id",
							RoomID: "room2-id",
							Name:   "test2",
						},
						"user3-id": &model.User{
							ID:     "user3-id",
							RoomID: "room2-id",
							Name:   "test3",
						},
					},
				}
				return r
			},
			roomID: "room2-id",
			expList: &storage.UserList{
				Items: []model.User{
					{
						ID:     "user2-id",
						RoomID: "room2-id",
						Name:   "test2",
					},
					{
						ID:     "user3-id",
						RoomID: "room2-id",
						Name:   "test3",
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			gotList, err := r.ListRoomUsers(context.TODO(), test.roomID)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				// Sort both for consistency.
				sort.Slice(test.expList.Items, func(i, j int) bool { return test.expList.Items[i].ID < test.expList.Items[j].ID })
				sort.Slice(gotList.Items, func(i, j int) bool { return gotList.Items[i].ID < gotList.Items[j].ID })
				assert.Equal(test.expList, gotList)
			}
		})
	}
}

func TestUserRepositoryGetUserByID(t *testing.T) {
	tests := map[string]struct {
		repo    func() *memory.UserRepository
		userID  string
		expUser model.User
		expErr  error
	}{
		"Having a user that does not exit, should return an error.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByID = map[string]*model.User{
					"user1-id": {},
				}
				return r
			},
			userID: "user2-id",
			expErr: internalerrors.ErrMissing,
		},

		"Having a user that exists in a room, should return the user.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByID = map[string]*model.User{
					"user1-id": {ID: "user1-id"},
				}
				return r
			},
			userID:  "user1-id",
			expUser: model.User{ID: "user1-id"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			gotUser, err := r.GetUserByID(context.TODO(), test.userID)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				assert.Equal(test.expUser, *gotUser)
			}
		})
	}
}

func TestUserRepositoryUserExists(t *testing.T) {
	tests := map[string]struct {
		repo      func() *memory.UserRepository
		userID    string
		expExists bool
		expErr    error
	}{
		"Having a user that does not exit, should return no exists.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByID = map[string]*model.User{
					"user1-id": {},
				}
				return r
			},
			userID:    "user2-id",
			expExists: false,
		},

		"Having a user that exists in a room, should return exists.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByID = map[string]*model.User{
					"user1-id": {},
				}
				return r
			},
			userID:    "user1-id",
			expExists: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			gotExists, err := r.UserExists(context.TODO(), test.userID)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				assert.Equal(test.expExists, gotExists)
			}
		})
	}
}

func TestUserRepositoryUserExistsByNameInsensitive(t *testing.T) {
	tests := map[string]struct {
		repo      func() *memory.UserRepository
		roomID    string
		username  string
		expExists bool
		expErr    error
	}{
		"Having a user user that is not in the room, should return no exists.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByRoom = map[string]map[string]*model.User{
					"room1-id": {
						"user1-id": &model.User{
							Name: "test1",
						},
					},
				}
				return r
			},
			roomID:    "room2-id",
			username:  "test1",
			expExists: false,
		},

		"Having a user user that matches exactly, should return exists.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByRoom = map[string]map[string]*model.User{
					"room1-id": {
						"user1-id": &model.User{
							Name: "test1",
						},
					},
				}
				return r
			},
			roomID:    "room1-id",
			username:  "test1",
			expExists: true,
		},

		"Having a user user that matches case insensitive, should return exists.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByRoom = map[string]map[string]*model.User{
					"room1-id": {
						"user1-id": &model.User{
							Name: "TeSt1",
						},
					},
				}
				return r
			},
			roomID:    "room1-id",
			username:  "tEsT1",
			expExists: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			gotExists, err := r.UserExistsByNameInsensitive(context.TODO(), test.roomID, test.username)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				assert.Equal(test.expExists, gotExists)
			}
		})
	}
}

func TestUserRepositoryGetUserByNameInsensitive(t *testing.T) {
	tests := map[string]struct {
		repo     func() *memory.UserRepository
		roomID   string
		username string
		expUser  model.User
		expErr   error
	}{
		"Having a user that is not in the room, should fail.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByRoom = map[string]map[string]*model.User{
					"room1-id": {
						"user1-id": &model.User{
							Name: "test1",
						},
					},
				}
				return r
			},
			roomID:   "room2-id",
			username: "test1",
			expErr:   internalerrors.ErrMissing,
		},

		"Having a user user that matches exactly, should return the user.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByRoom = map[string]map[string]*model.User{
					"room1-id": {
						"user1-id": &model.User{
							ID:   "user1-id",
							Name: "test1",
						},
					},
				}
				return r
			},
			roomID:   "room1-id",
			username: "test1",
			expUser:  model.User{ID: "user1-id", Name: "test1"},
		},

		"Having a user that matches case insensitive, should return the user.": {
			repo: func() *memory.UserRepository {
				r := memory.NewUserRepository()
				r.UsersByRoom = map[string]map[string]*model.User{
					"room1-id": {
						"user1-id": &model.User{
							ID:   "user1-id",
							Name: "TeSt1",
						},
					},
				}
				return r
			},
			roomID:   "room1-id",
			username: "tEsT1",
			expUser:  model.User{ID: "user1-id", Name: "TeSt1"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			gotUser, err := r.GetUserByNameInsensitive(context.TODO(), test.roomID, test.username)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				assert.Equal(test.expUser, *gotUser)
			}
		})
	}
}
