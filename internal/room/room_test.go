package room_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/room/roommock"
)

func TestServiceCreateRoom(t *testing.T) {
	tests := map[string]struct {
		config  room.ServiceConfig
		mock    func(r *roommock.Repository)
		req     func() room.CreateRoomRequest
		expResp func() *room.CreateRoomResponse
		expErr  bool
	}{
		"Having a creation request without name, should fail.": {
			mock: func(r *roommock.Repository) {},
			req: func() room.CreateRoomRequest {
				return room.CreateRoomRequest{Name: ""}
			},
			expErr: true,
		},

		"Having a correct room creation it should store the room.": {
			mock: func(r *roommock.Repository) {
				exp := model.Room{
					ID:   "test",
					Name: "test-room",
				}
				r.On("CreateRoom", mock.Anything, exp).Once().Return(nil)
			},
			req: func() room.CreateRoomRequest {
				return room.CreateRoomRequest{Name: "test-room"}
			},
			expResp: func() *room.CreateRoomResponse {
				return &room.CreateRoomResponse{
					Room: model.Room{
						ID:   "test",
						Name: "test-room",
					},
				}
			},
		},

		"Having a correct request and an error while storing, it should fail.": {
			mock: func(r *roommock.Repository) {
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
			mr := &roommock.Repository{}
			test.mock(mr)

			test.config.RoomRepository = mr
			test.config.IDGenerator = func() string { return "test" }

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
