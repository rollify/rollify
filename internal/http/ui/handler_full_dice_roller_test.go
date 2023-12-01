package ui_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/r3labs/sse/v2"
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

func TestHanderFullDiceRoller(t *testing.T) {
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
		"Entering on the room index should show the dice roller.": {
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/u/room/e02b402d-c23b-45b2-a5ea-583a566a9a6b", nil)
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				req.AddCookie(&http.Cookie{Name: "_room_user_id_e02b402d-c23b-45b2-a5ea-583a566a9a6b", Value: "user1", MaxAge: 999999999999})

				return req
			},
			mock: func(m mocks) {
				r := room.GetRoomRequest{ID: "e02b402d-c23b-45b2-a5ea-583a566a9a6b"}
				m.mr.On("GetRoom", mock.Anything, r).Once().Return(&room.GetRoomResponse{Room: model.Room{
					ID:   "e02b402d-c23b-45b2-a5ea-583a566a9a6b",
					Name: "test",
				}}, nil)
			},
			expHeaders: http.Header{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			expCode: 200,
			expBody: []string{
				`<a role="button" class="contrast" href="/u/room/e02b402d-c23b-45b2-a5ea-583a566a9a6b/dice-roll-history" hx-ext="sse" sse-connect="/u/subscribe/room/dice-roll-history?stream=notification-e02b402d-c23b-45b2-a5ea-583a566a9a6b" sse-swap="new_dice_roll" hx-swap="none"> Dice roll history <span id="notification-badge">0</span> </a>`, // We have the dice history button and bubble notification SSE connection with HTMX.
				`<a href="/u/logout/e02b402d-c23b-45b2-a5ea-583a566a9a6b" role="button" class="secondary outline"> Logout </a>`,                                   // We have the logout button.
				`<form id="diceRollerForm" hx-post="/u/room/e02b402d-c23b-45b2-a5ea-583a566a9a6b/new-dice-roll" hx-swap="innerHTML" hx-target="#diceRollResult">`, // Check HTMX call is in place.
				`<select id="d4" name="d4" class="diceRollerSelector">`,                                                                                           // We have a d4 on a the dice roller.
				`<select id="d6" name="d6" class="diceRollerSelector">`,                                                                                           // We have a d6 on a the dice roller.
				`<select id="d8" name="d8" class="diceRollerSelector">`,                                                                                           // We have a d8 on a the dice roller.
				`<select id="d10" name="d10" class="diceRollerSelector">`,                                                                                         // We have a d10 on a the dice roller.
				`<select id="d12" name="d12" class="diceRollerSelector">`,                                                                                         // We have a d12 on a the dice roller.
				`<select id="d20" name="d20" class="diceRollerSelector">`,                                                                                         // We have a d20 on a the dice roller.
				`<a onclick="cleanDiceSelectors()" href="#" role="button" class="secondary">Clear</a> </div> `,                                                    // We have the clear button.
				`<button type="submit">Roll</button>`,                                                                                                             // We have the submit button.
				`<footer id="diceRollResult"> <!-- will be replaced by HTMX on dice rolls--> </footer>`,                                                           // We have the empty result of the dice roll.
				`<nav class="container-fluid">`,                                                                                                                   // We have a nav bar.
				`<footer class="container-fluid">`,                                                                                                                // We have a footer.
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
