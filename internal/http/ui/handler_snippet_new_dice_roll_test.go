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

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/http/ui"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room/roommock"
	"github.com/rollify/rollify/internal/user/usermock"
)

func TestHandlerSnippetNewDiceRoll(t *testing.T) {
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
		"Creating a new dice roll should render the dice roll and return the result as an HTML HTMX snippet .": {
			request: func() *http.Request {
				form := url.Values{}
				form.Add("d4", "2")
				form.Add("d20", "1")
				req := httptest.NewRequest(http.MethodPost, "/u/room/e02b402d-c23b-45b2-a5ea-583a566a9a6b/new-dice-roll", strings.NewReader(form.Encode()))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				req.AddCookie(&http.Cookie{Name: "_room_user_id_e02b402d-c23b-45b2-a5ea-583a566a9a6b", Value: "user1", MaxAge: 999999999999})

				return req
			},
			mock: func(m mocks) {
				r := dice.CreateDiceRollRequest{UserID: "user1", RoomID: "e02b402d-c23b-45b2-a5ea-583a566a9a6b", Dice: []model.DieType{
					model.DieTypeD4,
					model.DieTypeD4,
					model.DieTypeD20,
				}}
				m.md.On("CreateDiceRoll", mock.Anything, r).Once().Return(&dice.CreateDiceRollResponse{DiceRoll: model.DiceRoll{
					ID: "test1",
					Dice: []model.DieRoll{
						{ID: "2", Type: model.DieTypeD4, Side: 2}, // Force unsorted to check sorted render HTML.
						{ID: "1", Type: model.DieTypeD4, Side: 1},
						{ID: "3", Type: model.DieTypeD20, Side: 3},
					},
				}}, nil)
			},
			expHeaders: http.Header{
				"Content-Type": {"text/plain; charset=utf-8"},
			},
			expCode: 200,
			expBody: []string{
				`<figure id="dice-roll-result">`, // We have the dice roll result.
				`<title>D4</title>`,              // We have d4 table header title.
				`<title>D20</title>`,             // We have d20 table header title.
				`<tr> <td> <kbd>1</kbd> <kbd>2</kbd> </td> <td> <kbd>3</kbd> </td> </tr>`, // We have all dice roll results (sorted) as a table row.
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
