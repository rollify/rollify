package apiv1_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/http/apiv1"
	"github.com/rollify/rollify/internal/model"
)

func TestAPIV1Pong(t *testing.T) {
	tests := map[string]struct {
		req           func() *http.Request
		expStatusCode int
		expBody       string
	}{
		"Having a correct request, should be handled correctly.": {
			req: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/ping", nil)
				return r
			},
			expStatusCode: http.StatusOK,
			expBody:       `"pong"`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Prepare.
			cfg := apiv1.Config{
				DiceAppService: &dicemock.Service{},
			}
			h, err := apiv1.New(cfg)
			require.NoError(err)

			// Execute.
			w := httptest.NewRecorder()
			h.ServeHTTP(w, test.req())

			// Check.
			res := w.Result()
			gotBody, err := ioutil.ReadAll(res.Body)
			require.NoError(err)
			assert.Equal(test.expStatusCode, res.StatusCode)
			assert.Equal(test.expBody, string(gotBody))
		})
	}
}

func TestAPIV1ListDiceTypes(t *testing.T) {
	tests := map[string]struct {
		mock          func(*dicemock.Service)
		req           func() *http.Request
		expStatusCode int
		expBody       string
	}{
		"Having a correct request, should be handled correctly.": {
			mock: func(m *dicemock.Service) {
				exp := &dice.ListDiceTypesResponse{
					DiceTypes: []model.DieType{model.DieTypeD4},
				}
				m.On("ListDiceTypes", mock.Anything).Once().Return(exp, nil)
			},
			req: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/types", nil)
				return r
			},
			expStatusCode: http.StatusOK,
			expBody: `{
 "items": [
  {
   "id": "d4",
   "name": "d4",
   "sides": 4
  }
 ]
}`,
		},

		"Having an internal error on the applicaiton service should return an internal error.": {
			mock: func(m *dicemock.Service) {
				m.On("ListDiceTypes", mock.Anything).Once().Return(nil, errors.New("wanted error"))
			},
			req: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/types", nil)
				return r
			},
			expStatusCode: http.StatusInternalServerError,
			expBody:       `wanted error`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			md := &dicemock.Service{}
			test.mock(md)

			// Prepare.
			cfg := apiv1.Config{
				DiceAppService: md,
			}
			h, err := apiv1.New(cfg)
			require.NoError(err)

			// Execute.
			w := httptest.NewRecorder()
			h.ServeHTTP(w, test.req())

			// Check.
			res := w.Result()
			gotBody, err := ioutil.ReadAll(res.Body)
			require.NoError(err)
			assert.Equal(test.expStatusCode, res.StatusCode)
			assert.Equal(test.expBody, string(gotBody))
		})
	}
}

func TestAPIV1CreateDiceRoll(t *testing.T) {
	tests := map[string]struct {
		mock          func(*dicemock.Service)
		req           func() *http.Request
		expStatusCode int
		expBody       string
	}{
		"Having a request without user ID should fail.": {
			mock: func(m *dicemock.Service) {},
			req: func() *http.Request {
				body := `{"user_id": "","room_id": "test-room", "dice_type_ids": ["d20"]}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/roll", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `user_id is required`,
		},

		"Having a request without room ID should fail.": {
			mock: func(m *dicemock.Service) {},
			req: func() *http.Request {
				body := `{"user_id": "test-user","room_id": "", "dice_type_ids": ["d20"]}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/roll", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `room_id is required`,
		},

		"Having a request without dice types should fail.": {
			mock: func(m *dicemock.Service) {},
			req: func() *http.Request {
				body := `{"user_id": "test-user","room_id": "test-room", "dice_type_ids": []}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/roll", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `dice_type_ids are required`,
		},

		"Having a request with invalid dice types should fail .": {
			mock: func(m *dicemock.Service) {},
			req: func() *http.Request {
				body := `{"user_id": "test-user","room_id": "test-room", "dice_type_ids": ["d99999"]}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/roll", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `d99999 die type is not valid`,
		},

		"Having a correct request that fails creating the roll should fail.": {
			mock: func(m *dicemock.Service) {
				m.On("CreateDiceRoll", mock.Anything, mock.Anything).Once().Return(nil, errors.New("wanted error"))
			},
			req: func() *http.Request {
				body := `{"user_id": "test-user","room_id": "test-room", "dice_type_ids": ["d6", "d20"]}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/roll", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusInternalServerError,
			expBody:       `wanted error`,
		},

		"Having a correct request should create the dice roll correctly.": {
			mock: func(m *dicemock.Service) {
				expReq := dice.CreateDiceRollRequest{
					UserID: "test-user",
					RoomID: "test-room",
					Dice:   []model.DieType{model.DieTypeD6, model.DieTypeD20},
				}
				resp := &dice.CreateDiceRollResponse{
					DiceRoll: model.DiceRoll{
						ID: "test-dice-roll",
						Dice: []model.DieRoll{
							{ID: "dice-1", Type: model.DieTypeD6, Side: 5},
							{ID: "dice-2", Type: model.DieTypeD20, Side: 18},
						},
					},
				}
				m.On("CreateDiceRoll", mock.Anything, expReq).Once().Return(resp, nil)
			},
			req: func() *http.Request {
				body := `{"user_id": "test-user","room_id": "test-room", "dice_type_ids": ["d6", "d20"]}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/roll", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusCreated,
			expBody: `{
 "id": "test-dice-roll",
 "dice": [
  {
   "id": "dice-1",
   "dice_type_id": "d6",
   "side": 5
  },
  {
   "id": "dice-2",
   "dice_type_id": "d20",
   "side": 18
  }
 ]
}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			md := &dicemock.Service{}
			test.mock(md)

			// Prepare.
			cfg := apiv1.Config{
				DiceAppService: md,
			}
			h, err := apiv1.New(cfg)
			require.NoError(err)

			// Execute.
			w := httptest.NewRecorder()
			h.ServeHTTP(w, test.req())

			// Check.
			res := w.Result()
			gotBody, err := ioutil.ReadAll(res.Body)
			require.NoError(err)
			assert.Equal(test.expStatusCode, res.StatusCode)
			assert.Equal(test.expBody, string(gotBody))
		})
	}
}
