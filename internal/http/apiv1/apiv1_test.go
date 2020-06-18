package apiv1_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/http/apiv1"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/room/roommock"
	"github.com/rollify/rollify/internal/user"
	"github.com/rollify/rollify/internal/user/usermock"
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
				RoomAppService: &roommock.Service{},
				UserAppService: &usermock.Service{},
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
				RoomAppService: &roommock.Service{},
				UserAppService: &usermock.Service{},
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
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")

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
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/rolls", strings.NewReader(body))
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
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/rolls", strings.NewReader(body))
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
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/rolls", strings.NewReader(body))
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
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/rolls", strings.NewReader(body))
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
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/rolls", strings.NewReader(body))
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
						ID:        "test-dice-roll",
						CreatedAt: t0,
						UserID:    "test-user",
						RoomID:    "test-room",
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
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/dice/rolls", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusCreated,
			expBody: `{
 "id": "test-dice-roll",
 "created_at": "1912-06-23T01:02:03Z",
 "room_id": "test-room",
 "user_id": "test-user",
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
				RoomAppService: &roommock.Service{},
				UserAppService: &usermock.Service{},
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

func TestAPIV1ListDiceRolls(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")

	tests := map[string]struct {
		mock          func(*dicemock.Service)
		req           func() *http.Request
		expStatusCode int
		expBody       string
	}{
		"Having a request without room id should fail.": {
			mock: func(m *dicemock.Service) {},
			req: func() *http.Request {
				body := `{"room_id": "", "user_id": "user-id"}`
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/rolls", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `room_id is required`,
		},

		"Having a request with an error form the app service, should fail.": {
			mock: func(m *dicemock.Service) {
				m.On("ListDiceRolls", mock.Anything, mock.Anything).Once().Return(nil, errors.New("wanted error"))
			},
			req: func() *http.Request {
				body := `{"room_id": "room-id", "user_id": "user-id"}`
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/rolls", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusInternalServerError,
			expBody:       `wanted error`,
		},

		"Having a request should return dice rolls.": {
			mock: func(m *dicemock.Service) {
				expReq := dice.ListDiceRollsRequest{RoomID: "room-id", UserID: "user-id"}
				resp := &dice.ListDiceRollsResponse{
					DiceRolls: []model.DiceRoll{
						{
							ID:        "dr1",
							CreatedAt: t0,
							UserID:    "user-1",
							RoomID:    "room-1",
							Dice: []model.DieRoll{
								{ID: "d1", Type: model.DieTypeD6, Side: 4},
								{ID: "d2", Type: model.DieTypeD6, Side: 5},
							},
						},
						{
							ID:        "dr2",
							CreatedAt: t0,
							UserID:    "user-2",
							RoomID:    "room-2",
							Dice: []model.DieRoll{
								{ID: "d3", Type: model.DieTypeD20, Side: 18},
							},
						},
					},
				}
				m.On("ListDiceRolls", mock.Anything, expReq).Once().Return(resp, nil)
			},
			req: func() *http.Request {
				body := `{"room_id": "room-id", "user_id": "user-id"}`
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/rolls", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusOK,
			expBody: `{
 "items": [
  {
   "id": "dr1",
   "created_at": "1912-06-23T01:02:03Z",
   "user_id": "user-1",
   "room_id": "room-1",
   "dice": [
    {
     "id": "d1",
     "type_id": "d6",
     "side": 4
    },
    {
     "id": "d2",
     "type_id": "d6",
     "side": 5
    }
   ]
  },
  {
   "id": "dr2",
   "created_at": "1912-06-23T01:02:03Z",
   "user_id": "user-2",
   "room_id": "room-2",
   "dice": [
    {
     "id": "d3",
     "type_id": "d20",
     "side": 18
    }
   ]
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
				RoomAppService: &roommock.Service{},
				UserAppService: &usermock.Service{},
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

func TestAPIV1CreateRoom(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")

	tests := map[string]struct {
		mock          func(*roommock.Service)
		req           func() *http.Request
		expStatusCode int
		expBody       string
	}{
		"Having a request without name should fail.": {
			mock: func(m *roommock.Service) {},
			req: func() *http.Request {
				body := `{"name": ""}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/rooms", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `name is required`,
		},

		"Having a correct request that fails creating the the room should fail.": {
			mock: func(m *roommock.Service) {
				m.On("CreateRoom", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("wanted error"))
			},
			req: func() *http.Request {
				body := `{"name": "test-room"}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/rooms", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusInternalServerError,
			expBody:       `wanted error`,
		},

		"Having a correct request should create the room.": {
			mock: func(m *roommock.Service) {
				exp := room.CreateRoomRequest{Name: "test-room"}
				resp := &room.CreateRoomResponse{Room: model.Room{
					Name:      "test-room",
					CreatedAt: t0,
					ID:        "room-id",
				}}
				m.On("CreateRoom", mock.Anything, exp).Once().Return(resp, nil)
			},
			req: func() *http.Request {
				body := `{"name": "test-room"}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/rooms", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusCreated,
			expBody: `{
 "id": "room-id",
 "created_at": "1912-06-23T01:02:03Z",
 "name": "test-room"
}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			mr := &roommock.Service{}
			test.mock(mr)

			// Prepare.
			cfg := apiv1.Config{
				DiceAppService: &dicemock.Service{},
				RoomAppService: mr,
				UserAppService: &usermock.Service{},
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

func TestAPIV1CreateUser(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")

	tests := map[string]struct {
		mock          func(*usermock.Service)
		req           func() *http.Request
		expStatusCode int
		expBody       string
	}{
		"Having a request without name should fail.": {
			mock: func(m *usermock.Service) {},
			req: func() *http.Request {
				body := `{"name": "", "room_id": "room1-id"}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `name is required`,
		},

		"Having a request without room id should fail.": {
			mock: func(m *usermock.Service) {},
			req: func() *http.Request {
				body := `{"name": "test1", "room_id": ""}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `room_id is required`,
		},

		"Having a correct request that fails creating the the user should fail.": {
			mock: func(m *usermock.Service) {
				m.On("CreateUser", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("wanted error"))
			},
			req: func() *http.Request {
				body := `{"name": "test-room", "room_id": "test1-id"}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusInternalServerError,
			expBody:       `wanted error`,
		},

		"Having a correct request should create the user.": {
			mock: func(m *usermock.Service) {
				exp := user.CreateUserRequest{Name: "test1", RoomID: "test1-id"}
				resp := &user.CreateUserResponse{User: model.User{
					ID:        "test1-id",
					RoomID:    "test1-id",
					Name:      "test1",
					CreatedAt: t0,
				}}
				m.On("CreateUser", mock.Anything, exp).Once().Return(resp, nil)
			},
			req: func() *http.Request {
				body := `{"name": "test1", "room_id": "test1-id"}`
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusCreated,
			expBody: `{
 "id": "test1-id",
 "created_at": "1912-06-23T01:02:03Z",
 "name": "test1",
 "room_id": "test1-id"
}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			mu := &usermock.Service{}
			test.mock(mu)

			// Prepare.
			cfg := apiv1.Config{
				DiceAppService: &dicemock.Service{},
				RoomAppService: &roommock.Service{},
				UserAppService: mu,
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
