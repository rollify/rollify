package ui_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/http/ui"
	"github.com/rollify/rollify/internal/room/roommock"
	"github.com/rollify/rollify/internal/user/usermock"
)

var trimSpaceMultilineRegexp = regexp.MustCompile(`(?m)(^\s+|\s+$)`)

func assertContainsHTTPResponseBody(t *testing.T, exp []string, resp *httptest.ResponseRecorder) {
	// Sanitize got HTML so we make easier to check content.
	got := resp.Body.String()
	got = trimSpaceMultilineRegexp.ReplaceAllString(got, "")
	got = strings.Replace(got, "\n", " ", -1)

	// Check each expected snippet.
	for _, e := range exp {
		assert.Contains(t, got, e)
	}
}

func TestHanderFullIndex(t *testing.T) {
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
		"Calling the index directly should return the main index template.": {
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/u", nil)
			},
			mock: func(m mocks) {},
			expHeaders: http.Header{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			expCode: 200,
			expBody: []string{
				`<h1 id="index-title">The online dice roller for role players</h1>`,                                          // Make sure we are on the index.
				`<form id="createRoomForm" hx-post="/u/create-room" hx-swap="outerHTML" hx-target="#createRoomFormSection">`, // Check HTMX call is in place.
				`<div id="createRoomFormSection">`,                                                    // HTMX swap Target.
				`<input type="text" name="roomName" id="roomName" placeholder="Room name" required/>`, // Check The form has the important correct fields.
				`<nav class="container-fluid">`,                                                       // We have a nav bar.
				`<footer class="container-fluid">`,                                                    // We have a footer.
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
