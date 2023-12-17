package user_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
	"github.com/rollify/rollify/internal/storage/storagemock"
	"github.com/rollify/rollify/internal/user"
)

func TestServiceCreateUser(t *testing.T) {
	t0 := time.Now().UTC()

	tests := map[string]struct {
		config  user.ServiceConfig
		mock    func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository)
		req     func() user.CreateUserRequest
		expResp func() *user.CreateUserResponse
		expErr  bool
	}{
		"Having a creation request without name, should fail.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {},
			req: func() user.CreateUserRequest {
				return user.CreateUserRequest{Name: "", RoomID: "test-room"}
			},
			expErr: true,
		},

		"Having a creation request without room id, should fail.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {},
			req: func() user.CreateUserRequest {
				return user.CreateUserRequest{Name: "us-e_r.n'ame 42", RoomID: ""}
			},
			expErr: true,
		},

		"Having a creation request with not valid user name, should fail.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {},
			req: func() user.CreateUserRequest {
				return user.CreateUserRequest{Name: "usernamé", RoomID: "room-id"}
			},
			expErr: true,
		},

		"Having a creation request with in a room that does not exists, should error.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, "room-id").Once().Return(false, nil)
			},
			req: func() user.CreateUserRequest {
				return user.CreateUserRequest{Name: "us-e_r.n'ame 42", RoomID: "room-id"}
			},
			expErr: true,
		},

		"Having a creation request with an error while checking database, should error.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, "room-id").Once().Return(false, errors.New("wanted error"))
			},
			req: func() user.CreateUserRequest {
				return user.CreateUserRequest{Name: "us-e_r.n'ame 42", RoomID: "room-id"}
			},
			expErr: true,
		},

		"Having a creation request with an user that already exists, should return the user.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				ru.On("GetUserByNameInsensitive", mock.Anything, "room-id", "us-e_r.n'ame 42").Once().Return(&model.User{
					ID:        "test",
					Name:      "us-e_r.n'ame 42",
					RoomID:    "room-id",
					CreatedAt: t0,
				}, nil)
			},
			req: func() user.CreateUserRequest {
				return user.CreateUserRequest{Name: "us-e_r.n'ame 42", RoomID: "room-id"}
			},
			expResp: func() *user.CreateUserResponse {
				return &user.CreateUserResponse{
					User: model.User{
						ID:        "test",
						Name:      "us-e_r.n'ame 42",
						RoomID:    "room-id",
						CreatedAt: t0,
					},
				}
			},
		},

		"Having a creation request with an error while checking the user that already exists, should error.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				ru.On("GetUserByNameInsensitive", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("something"))
			},
			req: func() user.CreateUserRequest {
				return user.CreateUserRequest{Name: "us-e_r.n'ame 42", RoomID: "room-id"}
			},
			expErr: true,
		},

		"Having a creation request with an user, should create the user.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				ru.On("GetUserByNameInsensitive", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, internalerrors.ErrMissing)
				expUser := model.User{
					ID:        "test",
					Name:      "us-e_r.n'ame 42",
					RoomID:    "room-id",
					CreatedAt: t0,
				}
				ru.On("CreateUser", mock.Anything, expUser).Once().Return(nil)
			},
			req: func() user.CreateUserRequest {
				return user.CreateUserRequest{Name: "us-e_r.n'ame 42", RoomID: "room-id"}
			},
			expResp: func() *user.CreateUserResponse {
				return &user.CreateUserResponse{
					User: model.User{
						ID:        "test",
						Name:      "us-e_r.n'ame 42",
						RoomID:    "room-id",
						CreatedAt: t0,
					},
				}
			},
		},

		"Having a creation request with an error while storing the user, should fail.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				ru.On("GetUserByNameInsensitive", mock.Anything, "room-id", "us-e_r.n'ame 42").Once().Return(nil, internalerrors.ErrMissing)
				ru.On("CreateUser", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("wanted error"))
			},
			req: func() user.CreateUserRequest {
				return user.CreateUserRequest{Name: "us-e_r.n'ame 42", RoomID: "room-id"}
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks
			mr := &storagemock.RoomRepository{}
			mu := &storagemock.UserRepository{}
			test.mock(mu, mr)

			test.config.RoomRepository = mr
			test.config.UserRepository = mu
			test.config.IDGenerator = func() string { return "test" }
			test.config.TimeNowFunc = func() time.Time { return t0 }

			svc, err := user.NewService(test.config)
			require.NoError(err)

			gotResp, err := svc.CreateUser(context.TODO(), test.req())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResp(), gotResp)
			}
		})
	}
}

func TestServiceListUsers(t *testing.T) {
	t0 := time.Now().UTC()

	tests := map[string]struct {
		config  user.ServiceConfig
		mock    func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository)
		req     func() user.ListUsersRequest
		expResp func() *user.ListUsersResponse
		expErr  bool
	}{
		"Having a list request without room ID, should fail.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {},
			req: func() user.ListUsersRequest {
				return user.ListUsersRequest{RoomID: ""}
			},
			expErr: true,
		},

		"Having a list request for a non existent room, should fail.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, "room-id").Once().Return(false, nil)
			},
			req: func() user.ListUsersRequest {
				return user.ListUsersRequest{RoomID: "room-id"}
			},
			expErr: true,
		},

		"Having a list request with an error while checking the room exists, should fail.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, mock.Anything).Once().Return(false, errors.New("wanted error"))
			},
			req: func() user.ListUsersRequest {
				return user.ListUsersRequest{RoomID: "room-id"}
			},
			expErr: true,
		},

		"Having a list request, should list the users.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, mock.Anything).Return(true, nil)
				users := &storage.UserList{
					Items: []model.User{
						{ID: "user1", Name: "username1"},
						{ID: "user2", Name: "username2"},
					},
				}
				ru.On("ListRoomUsers", mock.Anything, "room-id").Return(users, nil)
			},
			req: func() user.ListUsersRequest {
				return user.ListUsersRequest{RoomID: "room-id"}
			},
			expResp: func() *user.ListUsersResponse {
				return &user.ListUsersResponse{
					Users: []model.User{
						{ID: "user1", Name: "username1"},
						{ID: "user2", Name: "username2"},
					},
				}
			},
		},

		"Having a list request with an error while listing the users, should fail.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, mock.Anything).Return(true, nil)
				ru.On("ListRoomUsers", mock.Anything, mock.Anything).Return(nil, errors.New("wanted error"))
			},
			req: func() user.ListUsersRequest {
				return user.ListUsersRequest{RoomID: "room-id"}
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks
			mr := &storagemock.RoomRepository{}
			mu := &storagemock.UserRepository{}
			test.mock(mu, mr)

			test.config.RoomRepository = mr
			test.config.UserRepository = mu
			test.config.IDGenerator = func() string { return "test" }
			test.config.TimeNowFunc = func() time.Time { return t0 }

			svc, err := user.NewService(test.config)
			require.NoError(err)

			gotResp, err := svc.ListUsers(context.TODO(), test.req())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResp(), gotResp)
			}
		})
	}
}

func TestServiceGetUser(t *testing.T) {
	t0 := time.Now().UTC()

	tests := map[string]struct {
		config  user.ServiceConfig
		mock    func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository)
		req     func() user.GetUserRequest
		expResp func() *user.GetUserResponse
		expErr  bool
	}{
		"A missing user id should fail.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
			},
			req: func() user.GetUserRequest {
				return user.GetUserRequest{UserID: ""}
			},
			expErr: true,
		},

		"Getting the user should get the user correctly.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				ru.On("GetUserByID", mock.Anything, "user-id").Return(&model.User{
					ID:        "user-id",
					Name:      "User0",
					RoomID:    "room0",
					CreatedAt: t0,
				}, nil)
			},
			req: func() user.GetUserRequest {
				return user.GetUserRequest{UserID: "user-id"}
			},
			expResp: func() *user.GetUserResponse {
				return &user.GetUserResponse{
					User: model.User{
						ID:        "user-id",
						Name:      "User0",
						RoomID:    "room0",
						CreatedAt: t0,
					},
				}
			},
		},

		"Having an error while getting the user, should fail.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				ru.On("GetUserByID", mock.Anything, mock.Anything).Return(nil, errors.New("wanted error"))
			},
			req: func() user.GetUserRequest {
				return user.GetUserRequest{UserID: "user-id"}
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks
			mr := &storagemock.RoomRepository{}
			mu := &storagemock.UserRepository{}
			test.mock(mu, mr)

			test.config.RoomRepository = mr
			test.config.UserRepository = mu
			test.config.IDGenerator = func() string { return "test" }
			test.config.TimeNowFunc = func() time.Time { return t0 }

			svc, err := user.NewService(test.config)
			require.NoError(err)

			gotResp, err := svc.GetUser(context.TODO(), test.req())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResp(), gotResp)
			}
		})
	}
}
