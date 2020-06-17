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

	"github.com/rollify/rollify/internal/model"
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

		"Having a creation request with an user that already exists, should error.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				ru.On("UserExistsByNameInsensitive", mock.Anything, "room-id", "us-e_r.n'ame 42").Once().Return(true, nil)
			},
			req: func() user.CreateUserRequest {
				return user.CreateUserRequest{Name: "us-e_r.n'ame 42", RoomID: "room-id"}
			},
			expErr: true,
		},

		"Having a creation request with an error while checking the user that already exists, should error.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				ru.On("UserExistsByNameInsensitive", mock.Anything, mock.Anything, mock.Anything).Once().Return(false, errors.New("wanted error"))
			},
			req: func() user.CreateUserRequest {
				return user.CreateUserRequest{Name: "us-e_r.n'ame 42", RoomID: "room-id"}
			},
			expErr: true,
		},

		"Having a creation request with an user, should create the user.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				ru.On("UserExistsByNameInsensitive", mock.Anything, mock.Anything, mock.Anything).Once().Return(false, nil)
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

		"Having a creation request with an erroo while storing the user, should create the user.": {
			mock: func(ru *storagemock.UserRepository, rr *storagemock.RoomRepository) {
				rr.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				ru.On("UserExistsByNameInsensitive", mock.Anything, mock.Anything, mock.Anything).Once().Return(false, nil)
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
