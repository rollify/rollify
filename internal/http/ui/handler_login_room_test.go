package ui_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/http/ui"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/room/roommock"
	"github.com/rollify/rollify/internal/user"
	"github.com/rollify/rollify/internal/user/usermock"
)

func TestHanderLoginRoom(t *testing.T) {
	type mocks struct {
		md *dicemock.Service
		mr *roommock.Service
		mu *usermock.Service
	}

	tests := map[string]struct {
		request    func() *http.Request
		mock       func(m mocks)
		expBody    string
		expHeaders http.Header
		expCode    int
	}{
		"Calling the login room without logged users should return the login template with no users.": {
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/u/login-room/e02b402d-c23b-45b2-a5ea-583a566a9a6b", nil)
			},
			mock: func(m mocks) {
				rgr := room.GetRoomRequest{ID: "e02b402d-c23b-45b2-a5ea-583a566a9a6b"}
				m.mr.On("GetRoom", mock.Anything, rgr).Once().Return(&room.GetRoomResponse{Room: model.Room{
					ID:   "e02b402d-c23b-45b2-a5ea-583a566a9a6b",
					Name: "test1",
				}}, nil)

				rlu := user.ListUsersRequest{RoomID: "e02b402d-c23b-45b2-a5ea-583a566a9a6b"}
				m.mu.On("ListUsers", mock.Anything, rlu).Once().Return(&user.ListUsersResponse{Users: []model.User{}}, nil)
			},
			expHeaders: http.Header{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			expCode: 200,
			expBody: "",
		},

		"Calling the login room with already logged users should return the login template with the current users loaded.": {
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/u/login-room/e02b402d-c23b-45b2-a5ea-583a566a9a6b", nil)
			},
			mock: func(m mocks) {
				rgr := room.GetRoomRequest{ID: "e02b402d-c23b-45b2-a5ea-583a566a9a6b"}
				m.mr.On("GetRoom", mock.Anything, rgr).Once().Return(&room.GetRoomResponse{Room: model.Room{
					ID:   "e02b402d-c23b-45b2-a5ea-583a566a9a6b",
					Name: "test1",
				}}, nil)

				rlu := user.ListUsersRequest{RoomID: "e02b402d-c23b-45b2-a5ea-583a566a9a6b"}
				m.mu.On("ListUsers", mock.Anything, rlu).Once().Return(&user.ListUsersResponse{Users: []model.User{
					{ID: "user1", Name: "User 1"},
					{ID: "user2", Name: "User 2"},
					{ID: "user3", Name: "User 3"},
				}}, nil)
			},
			expHeaders: http.Header{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			expCode: 200,
			expBody: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			m := mocks{
				md: &dicemock.Service{},
				mr: &roommock.Service{},
				mu: &usermock.Service{},
			}
			test.mock(m)

			h, err := ui.New(ui.Config{
				DiceAppService: m.md,
				RoomAppService: m.mr,
				UserAppService: m.mu,
			})
			require.NoError(err)

			w := httptest.NewRecorder()
			h.ServeHTTP(w, test.request())

			assert.Equal(test.expCode, w.Code)
			assert.Equal(test.expHeaders, w.Header())
			// TODO(slok).
			//assert.Equal(test.expBody, w.Body.String())
		})
	}
}
