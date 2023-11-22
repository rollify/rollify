package ui_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/http/ui"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/room/roommock"
	"github.com/rollify/rollify/internal/user/usermock"
)

func TestHanderCreateRoom(t *testing.T) {
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
		"Creating a new room, should create the room.": {
			request: func() *http.Request {
				form := url.Values{}
				form.Add("roomName", "test1")
				req := httptest.NewRequest(http.MethodPost, "/u/create-room", strings.NewReader(form.Encode()))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			mock: func(m mocks) {
				rgr := room.CreateRoomRequest{Name: "test1"}
				m.mr.On("CreateRoom", mock.Anything, rgr).Once().Return(&room.CreateRoomResponse{Room: model.Room{
					ID:   "e02b402d-c23b-45b2-a5ea-583a566a9a6b",
					Name: "test1",
				}}, nil)

			},
			expHeaders: http.Header{
				"Hx-Redirect": {"/u/login-room/e02b402d-c23b-45b2-a5ea-583a566a9a6b"},
			},
			expCode: 200,
			expBody: "",
		},

		"An empty room name should error.": {
			request: func() *http.Request {
				form := url.Values{}
				form.Add("roomName", "      ")
				req := httptest.NewRequest(http.MethodPost, "/u/create-room", strings.NewReader(form.Encode()))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			mock: func(m mocks) {
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
