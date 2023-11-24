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
	"github.com/rollify/rollify/internal/user/usermock"
)

func TestHanderdiceRollHistory(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "2023-01-21T11:05:45Z")
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
		"Asking for the dice roll history items should return the page with the list.": {
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
					UserID:   "user1",
					RoomID:   "e02b402d-c23b-45b2-a5ea-583a566a9a6b",
					PageOpts: model.PaginationOpts{Size: 10},
				}
				m.md.On("ListDiceRolls", mock.Anything, r2).Once().Return(&dice.ListDiceRollsResponse{
					DiceRolls: []model.DiceRoll{},
				}, nil)
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
			// TODO(slok).
			//assert.Equal(test.expBody, w.Body.String())
		})
	}
}
