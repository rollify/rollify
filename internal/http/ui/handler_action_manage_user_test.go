package ui_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/http/ui"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room/roommock"
	"github.com/rollify/rollify/internal/user"
	"github.com/rollify/rollify/internal/user/usermock"
)

func TestHandlerActionManageUser(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "2023-01-21T11:05:45Z")
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
		"Creating a new user without data, should fail.": {
			request: func() *http.Request {
				form := url.Values{}
				req := httptest.NewRequest(http.MethodPost, "/u/login/e02b402d-c23b-45b2-a5ea-583a566a9a6b/manage-user", strings.NewReader(form.Encode()))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Add("HX-Request", "true")
				return req
			},
			mock:       func(m mocks) {},
			expHeaders: http.Header{},
			expCode:    500,
			expBody:    []string{},
		},

		"Creating a new user should create a new user and redirect the to the room.": {
			request: func() *http.Request {
				form := url.Values{}
				form.Add("username", "user1")
				req := httptest.NewRequest(http.MethodPost, "/u/login/e02b402d-c23b-45b2-a5ea-583a566a9a6b/manage-user", strings.NewReader(form.Encode()))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Add("HX-Request", "true")
				return req
			},
			mock: func(m mocks) {
				r := user.CreateUserRequest{Name: "user1", RoomID: "e02b402d-c23b-45b2-a5ea-583a566a9a6b"}
				m.mu.On("CreateUser", mock.Anything, r).Once().Return(&user.CreateUserResponse{User: model.User{
					ID:   "u1",
					Name: "user1",
				}}, nil)
			},
			expHeaders: http.Header{
				"Hx-Redirect": {"/u/room/e02b402d-c23b-45b2-a5ea-583a566a9a6b"},
				"Set-Cookie":  {"_room_user_id_e02b402d-c23b-45b2-a5ea-583a566a9a6b=u1; Path=/; Expires=Sat, 04 Feb 2023 11:05:45 GMT"},
			},
			expCode: 200,
			expBody: []string{},
		},

		"Creating a new user that already exists should not fail and use the existing user instead.": {
			request: func() *http.Request {
				form := url.Values{}
				form.Add("username", "user1")
				req := httptest.NewRequest(http.MethodPost, "/u/login/e02b402d-c23b-45b2-a5ea-583a566a9a6b/manage-user", strings.NewReader(form.Encode()))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Add("HX-Request", "true")
				return req
			},
			mock: func(m mocks) {
				r := user.CreateUserRequest{Name: "user1", RoomID: "e02b402d-c23b-45b2-a5ea-583a566a9a6b"}
				m.mu.On("CreateUser", mock.Anything, r).Once().Return(&user.CreateUserResponse{User: model.User{
					ID:   "u1",
					Name: "user1",
				}}, nil)
			},
			expHeaders: http.Header{
				"Hx-Redirect": {"/u/room/e02b402d-c23b-45b2-a5ea-583a566a9a6b"},
				"Set-Cookie":  {"_room_user_id_e02b402d-c23b-45b2-a5ea-583a566a9a6b=u1; Path=/; Expires=Sat, 04 Feb 2023 11:05:45 GMT"},
			},
			expCode: 200,
			expBody: []string{},
		},

		"Using an existing user should select a the user and redirect the to the room.": {
			request: func() *http.Request {
				form := url.Values{}
				form.Add("userID", "12345")
				req := httptest.NewRequest(http.MethodPost, "/u/login/e02b402d-c23b-45b2-a5ea-583a566a9a6b/manage-user", strings.NewReader(form.Encode()))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Add("HX-Request", "true")
				return req
			},
			mock: func(m mocks) {
				r := user.GetUserRequest{UserID: "12345"}
				m.mu.On("GetUser", mock.Anything, r).Once().Return(&user.GetUserResponse{User: model.User{
					ID:   "12345",
					Name: "user1",
				}}, nil)
			},
			expHeaders: http.Header{
				"Hx-Redirect": {"/u/room/e02b402d-c23b-45b2-a5ea-583a566a9a6b"},
				"Set-Cookie":  {"_room_user_id_e02b402d-c23b-45b2-a5ea-583a566a9a6b=12345; Path=/; Expires=Sat, 04 Feb 2023 11:05:45 GMT"},
			},
			expCode: 200,
			expBody: []string{},
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
				TimeNow:        func() time.Time { return t0.UTC() },
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
