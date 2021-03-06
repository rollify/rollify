package room_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/storage/storagemock"
)

func TestServiceCreateRoom(t *testing.T) {
	t0 := time.Now().UTC()

	tests := map[string]struct {
		config  room.ServiceConfig
		mock    func(r *storagemock.RoomRepository)
		req     func() room.CreateRoomRequest
		expResp func() *room.CreateRoomResponse
		expErr  bool
	}{
		"Having a creation request without name, should fail.": {
			mock: func(r *storagemock.RoomRepository) {},
			req: func() room.CreateRoomRequest {
				return room.CreateRoomRequest{Name: ""}
			},
			expErr: true,
		},

		"Having a correct room creation it should store the room.": {
			mock: func(r *storagemock.RoomRepository) {
				exp := model.Room{
					ID:        "test",
					CreatedAt: t0,
					Name:      "test-room",
				}
				r.On("CreateRoom", mock.Anything, exp).Once().Return(nil)
			},
			req: func() room.CreateRoomRequest {
				return room.CreateRoomRequest{Name: "test-room"}
			},
			expResp: func() *room.CreateRoomResponse {
				return &room.CreateRoomResponse{
					Room: model.Room{
						ID:        "test",
						CreatedAt: t0,
						Name:      "test-room",
					},
				}
			},
		},

		"Having a correct request and an error while storing, it should fail.": {
			mock: func(r *storagemock.RoomRepository) {
				r.On("CreateRoom", mock.Anything, mock.Anything).Once().Return(errors.New("wanted error"))
			},
			req: func() room.CreateRoomRequest {
				return room.CreateRoomRequest{Name: "test-room"}
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
			test.mock(mr)

			test.config.RoomRepository = mr
			test.config.IDGenerator = func() string { return "test" }
			test.config.TimeNowFunc = func() time.Time { return t0 }

			svc, err := room.NewService(test.config)
			require.NoError(err)

			gotResp, err := svc.CreateRoom(context.TODO(), test.req())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResp(), gotResp)
			}
		})
	}
}

func TestServiceGetRoom(t *testing.T) {
	tests := map[string]struct {
		config  room.ServiceConfig
		mock    func(r *storagemock.RoomRepository)
		req     func() room.GetRoomRequest
		expResp func() *room.GetRoomResponse
		expErr  bool
	}{
		"Having a get request without id, should fail.": {
			mock: func(r *storagemock.RoomRepository) {},
			req: func() room.GetRoomRequest {
				return room.GetRoomRequest{ID: ""}
			},
			expErr: true,
		},

		"Having a get request without a correct room, should return the room.": {
			mock: func(r *storagemock.RoomRepository) {
				room := &model.Room{
					ID:   "test",
					Name: "test-room",
				}
				r.On("GetRoom", mock.Anything, "test").Once().Return(room, nil)
			},
			req: func() room.GetRoomRequest {
				return room.GetRoomRequest{ID: "test"}
			},
			expResp: func() *room.GetRoomResponse {
				return &room.GetRoomResponse{
					Room: model.Room{
						ID:   "test",
						Name: "test-room",
					},
				}
			},
		},

		"Having a get request with an error from the repository, should fail.": {
			mock: func(r *storagemock.RoomRepository) {
				r.On("GetRoom", mock.Anything, mock.Anything).Once().Return(nil, errors.New("wanted error"))
			},
			req: func() room.GetRoomRequest {
				return room.GetRoomRequest{ID: "test"}
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
			test.mock(mr)

			test.config.RoomRepository = mr
			svc, err := room.NewService(test.config)
			require.NoError(err)

			gotResp, err := svc.GetRoom(context.TODO(), test.req())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResp(), gotResp)
			}
		})
	}
}
