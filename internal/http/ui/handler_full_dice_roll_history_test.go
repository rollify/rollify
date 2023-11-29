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

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/http/ui"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/room/roommock"
	"github.com/rollify/rollify/internal/user"
	"github.com/rollify/rollify/internal/user/usermock"
)

func TestHandlerFullDiceRollHistory(t *testing.T) {
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
		"Asking for the dice roll history items should return the page with the list and have pagination in place when there is a cursor.": {
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/u/room/e02b402d-c23b-45b2-a5ea-583a566a9a6b/dice-roll-history", nil)
				req.AddCookie(&http.Cookie{Name: "_room_user_id_e02b402d-c23b-45b2-a5ea-583a566a9a6b", Value: "user1", MaxAge: 999999999999})

				return req
			},
			mock: func(m mocks) {
				r1 := room.GetRoomRequest{ID: "e02b402d-c23b-45b2-a5ea-583a566a9a6b"}
				m.mr.On("GetRoom", mock.Anything, r1).Once().Return(&room.GetRoomResponse{Room: model.Room{
					ID:   "e02b402d-c23b-45b2-a5ea-583a566a9a6b",
					Name: "test",
				}}, nil)

				r2 := dice.ListDiceRollsRequest{
					RoomID:   "e02b402d-c23b-45b2-a5ea-583a566a9a6b",
					PageOpts: model.PaginationOpts{Size: 10},
				}
				m.md.On("ListDiceRolls", mock.Anything, r2).Once().Return(&dice.ListDiceRollsResponse{
					Cursors: model.PaginationCursors{
						HasNext:    true,
						LastCursor: "cursor12345",
					},
					DiceRolls: []model.DiceRoll{
						{
							UserID:    "user-id1",
							CreatedAt: t0.Add(-5 * time.Second),
							Dice: []model.DieRoll{
								{ID: "1", Type: model.DieTypeD4, Side: 1},
								{ID: "2", Type: model.DieTypeD4, Side: 2},
								{ID: "3", Type: model.DieTypeD20, Side: 3},
							},
						},
						{
							UserID:    "user-id2",
							CreatedAt: t0.Add(-10 * time.Second),
							Dice: []model.DieRoll{
								{ID: "4", Type: model.DieTypeD6, Side: 4},
								{ID: "5", Type: model.DieTypeD10, Side: 8},
								{ID: "6", Type: model.DieTypeD12, Side: 11},
							},
						},
						{
							UserID:    "user-id3",
							CreatedAt: t0.Add(-40 * time.Second),
							Dice: []model.DieRoll{
								{ID: "7", Type: model.DieTypeD8, Side: 6},
								{ID: "8", Type: model.DieTypeD20, Side: 20}, // Force sort.
								{ID: "9", Type: model.DieTypeD20, Side: 1},
							},
						},
					},
				}, nil)

				r3 := user.ListUsersRequest{RoomID: "e02b402d-c23b-45b2-a5ea-583a566a9a6b"}
				m.mu.On("ListUsers", mock.Anything, r3).Once().Return(&user.ListUsersResponse{
					Users: []model.User{
						{ID: "user-id1", Name: "user1"},
						{ID: "user-id2", Name: "user2"},
						{ID: "user-id3", Name: "user3"},
					},
				}, nil)
			},
			expHeaders: http.Header{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			expCode: 200,
			expBody: []string{
				`<a href="/u/room/e02b402d-c23b-45b2-a5ea-583a566a9a6b" role="button">Roll dice!</a>`,                                                                                                                           // We have the roll dice button on the nav var.
				`<table role="grid" hx-ext="sse" sse-connect="/u/subscribe/room/dice-roll-history?stream=html-e02b402d-c23b-45b2-a5ea-583a566a9a6b" sse-swap="new_dice_roll" hx-target="#dice-roll-rows" hx-swap="afterbegin">`, // We have push updates using SSE notifications to update the table with the latest dice rolls.
				`<th></th>`,          // We have metadata header on dice roll history table.
				`<title>D4</title>`,  // We have d4 header on dice roll history table.
				`<title>D6</title>`,  // We have d6 header on dice roll history table.
				`<title>D8</title>`,  // We have d8 header on dice roll history table.
				`<title>D10</title>`, // We have d10 header on dice roll history table.
				`<title>D12</title>`, // We have d12 header on dice roll history table.
				`<title>D20</title>`, // We have d20 header on dice roll history table.
				`<tr id="history-dice-roll-row"><td> <div> <strong>user1</strong> </div> <div> <small class="timestamp-ago" unix-ts="1674299140"></small> </div> </td> <td> <kbd>1</kbd> <kbd>2</kbd> </td> <td> </td> <td> </td> <td> </td> <td> </td> <td> <kbd>3</kbd> </td> </tr>`,                                                                                                                                                 // We have the results of 1st Dice roll.
				`<tr id="history-dice-roll-row"><td> <div> <strong>user2</strong> </div> <div> <small class="timestamp-ago" unix-ts="1674299135"></small> </div> </td> <td> </td> <td> <kbd>4</kbd> </td> <td> </td> <td> <kbd>8</kbd> </td> <td> <kbd>11</kbd> </td> <td> </td> </tr>`,                                                                                                                                                // We have the results of 2nd Dice roll.
				`<tr id="history-dice-roll-row" hx-trigger="revealed" hx-get="/u/room/e02b402d-c23b-45b2-a5ea-583a566a9a6b/dice-roll-history/more-items?cursor=cursor12345" hx-swap="afterend"><td> <div> <strong>user3</strong> </div> <div> <small class="timestamp-ago" unix-ts="1674299105"></small> </div> </td> <td> </td> <td> </td> <td> <kbd>6</kbd> </td> <td> </td> <td> </td> <td> <kbd>1</kbd> <kbd>20</kbd> </td> </tr>`, // We have the results of last Dice roll ans has HTMX infinite scroll calls.
				`<nav class="container-fluid">`,    // We have a nav bar.
				`<footer class="container-fluid">`, // We have a footer.
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
