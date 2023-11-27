package ui_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/r3labs/sse/v2"
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

func TestHanderFullLogin(t *testing.T) {
	type mocks struct {
		md *dicemock.Service
		mr *roommock.Service
		mu *usermock.Service
	}

	tests := map[string]struct {
		request    func() *http.Request
		mock       func(m mocks)
		expBody    []string
		expHeaders http.Header
		expCode    int
	}{
		"Calling the login room without logged users should return the login template with no users.": {
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/u/login/e02b402d-c23b-45b2-a5ea-583a566a9a6b", nil)
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
			expBody: []string{
				`<h1>Log in room "test1" </h1>`, // Make sure we are on the login page and the correct room.
				`<form id="LoginForm" hx-post="/u/login/e02b402d-c23b-45b2-a5ea-583a566a9a6b/manage-user" hx-swap="outerHTML"`, // Check HTMX call is in place.
				`<input type="text" name="username" id="username" placeholder="Username"/>`,                                    // Check The form has the important correct fields.
				`<nav class="container-fluid">`,    // We have a nav bar.
				`<footer class="container-fluid">`, // We have a footer.
			},
		},

		"Calling the login room with already logged users should return the login template with the current users loaded.": {
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/u/login/e02b402d-c23b-45b2-a5ea-583a566a9a6b", nil)
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
			expBody: []string{
				`<h1>Log in room "test1" </h1>`, // Make sure we are on the login page and the correct room.
				`<form id="LoginForm" hx-post="/u/login/e02b402d-c23b-45b2-a5ea-583a566a9a6b/manage-user" hx-swap="outerHTML"`, // Check HTMX call is in place.
				`<input type="text" name="username" id="username" placeholder="Username"/>`,                                    // Check The form has the important correct fields.
				`<nav class="container-fluid">`,    // We have a nav bar.
				`<footer class="container-fluid">`, // We have a footer.
				`<h4>Existing user</h4>`,           // We have existing users form part.
				`<select id="userID" name="userID"> <option value="" disabled selected>Select</option> <option value="user1">User 1</option> <option value="user2">User 2</option> <option value="user3">User 3</option> </select>`, // Existing users are selectable.
			},
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

			s := sse.New()
			defer s.Close()
			h, err := ui.New(ui.Config{
				DiceAppService: m.md,
				RoomAppService: m.mr,
				UserAppService: m.mu,
				SSEServer:      s,
			})
			require.NoError(err)

			w := httptest.NewRecorder()
			h.ServeHTTP(w, test.request())

			assert.Equal(test.expCode, w.Code)
			assert.Equal(test.expHeaders, w.Header())
			assertContainsHTTPResponseBody(t, test.expBody, w)
		})
	}
}
