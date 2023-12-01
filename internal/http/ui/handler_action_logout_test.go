package ui_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/http/ui"
	"github.com/rollify/rollify/internal/room/roommock"
	"github.com/rollify/rollify/internal/user/usermock"
)

func TestHandlerActionLogout(t *testing.T) {
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
		"On logout user should be deleted from cookies and redirect to the login of the room.": {
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/u/logout/e02b402d-c23b-45b2-a5ea-583a566a9a6b", nil)
				return req

			},
			mock: func(m mocks) {},
			expHeaders: http.Header{
				"Content-Type": {"text/html; charset=utf-8"},
				"Location":     {"/u/login/e02b402d-c23b-45b2-a5ea-583a566a9a6b"},
			},
			expCode: 307,
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
