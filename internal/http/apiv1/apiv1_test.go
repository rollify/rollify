package apiv1_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"nhooyr.io/websocket"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/http/apiv1"
	"github.com/rollify/rollify/internal/internalerrors"
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
			gotBody, err := io.ReadAll(res.Body)
			require.NoError(err)
			assert.Equal(test.expStatusCode, res.StatusCode)
			assert.Equal(test.expBody, string(gotBody))
		})
	}
}

// TesTestAPIV1ErrorMappings using one of the HTTP endpoints, checks if the
// internalerror error mappings to HTTP status codes are correct. We should check
// all possible errors on every handler, but this adds a lot of complexity, so
// at least we check the mappings are correct.
func TestAPIV1ErrorMappings(t *testing.T) {
	tests := map[string]struct {
		mock          func(*dicemock.Service)
		req           func() *http.Request
		expStatusCode int
		expBody       string
	}{
		"Having no ErrNotValid, should return 400.": {
			mock: func(m *dicemock.Service) {
				r := &dice.ListDiceTypesResponse{}
				m.On("ListDiceTypes", mock.Anything).Once().Return(r, fmt.Errorf("wanted error: %w", internalerrors.ErrNotValid))
			},
			req: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/types", nil)
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       "{\n \"Code\": 400,\n \"Message\": \"wanted error: not valid\",\n \"Header\": null\n}",
		},

		"Having no ErrMissing, should return 404.": {
			mock: func(m *dicemock.Service) {
				r := &dice.ListDiceTypesResponse{}
				m.On("ListDiceTypes", mock.Anything).Once().Return(r, fmt.Errorf("wanted error: %w", internalerrors.ErrMissing))
			},
			req: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/types", nil)
				return r
			},
			expStatusCode: http.StatusNotFound,
			expBody:       "{\n \"Code\": 404,\n \"Message\": \"wanted error: is missing\",\n \"Header\": null\n}",
		},

		"Having no ErrAlreadyExists, should return 409.": {
			mock: func(m *dicemock.Service) {
				r := &dice.ListDiceTypesResponse{}
				m.On("ListDiceTypes", mock.Anything).Once().Return(r, fmt.Errorf("wanted error: %w", internalerrors.ErrAlreadyExists))
			},
			req: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/types", nil)
				return r
			},
			expStatusCode: http.StatusConflict,
			expBody:       "{\n \"Code\": 409,\n \"Message\": \"wanted error: already exists\",\n \"Header\": null\n}",
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
			gotBody, err := io.ReadAll(res.Body)
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

		"Having an internal error on the application service should return an internal error.": {
			mock: func(m *dicemock.Service) {
				m.On("ListDiceTypes", mock.Anything).Once().Return(nil, errors.New("wanted error"))
			},
			req: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/types", nil)
				return r
			},
			expStatusCode: http.StatusInternalServerError,
			expBody:       "{\n \"Code\": 500,\n \"Message\": \"wanted error\",\n \"Header\": null\n}",
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
			gotBody, err := io.ReadAll(res.Body)
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
			expBody:       "{\n \"Code\": 400,\n \"Message\": \"user_id is required\",\n \"Header\": null\n}",
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
			expBody:       "{\n \"Code\": 400,\n \"Message\": \"room_id is required\",\n \"Header\": null\n}",
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
			expBody:       "{\n \"Code\": 400,\n \"Message\": \"dice_type_ids are required\",\n \"Header\": null\n}",
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
			expBody:       "{\n \"Code\": 400,\n \"Message\": \"d99999 die type is not valid\",\n \"Header\": null\n}",
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
			expBody:       "{\n \"Code\": 500,\n \"Message\": \"wanted error\",\n \"Header\": null\n}",
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
			gotBody, err := io.ReadAll(res.Body)
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
				q := url.Values{}
				q.Add("room-id", "")
				q.Add("user-id", "user-id")
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/rolls", nil)
				r.URL.RawQuery = q.Encode()
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       "{\n \"Code\": 400,\n \"Message\": \"room-id is required\",\n \"Header\": null\n}",
		},

		"Having a wrong order query should fail.": {
			mock: func(m *dicemock.Service) {},
			req: func() *http.Request {
				q := url.Values{}
				q.Add("room-id", "room-id")
				q.Add("user-id", "user-id")
				q.Add("order", "wrong")
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/rolls", nil)
				r.URL.RawQuery = q.Encode()
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       "{\n \"Code\": 400,\n \"Message\": \"pagination order 'wrong' is invalid\",\n \"Header\": null\n}",
		},

		"Having a request with an error form the app service, should fail.": {
			mock: func(m *dicemock.Service) {
				m.On("ListDiceRolls", mock.Anything, mock.Anything).Once().Return(nil, errors.New("wanted error"))
			},
			req: func() *http.Request {
				q := url.Values{}
				q.Add("room-id", "room-id")
				q.Add("user-id", "user-id")
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/rolls", nil)
				r.URL.RawQuery = q.Encode()
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusInternalServerError,
			expBody:       "{\n \"Code\": 500,\n \"Message\": \"wanted error\",\n \"Header\": null\n}",
		},

		"Having a request should return dice rolls.": {
			mock: func(m *dicemock.Service) {
				expReq := dice.ListDiceRollsRequest{
					RoomID: "room-id",
					UserID: "user-id",
					PageOpts: model.PaginationOpts{
						Cursor: "threepwood",
						Order:  model.PaginationOrderAsc,
					},
				}
				resp := &dice.ListDiceRollsResponse{
					Cursors: model.PaginationCursors{
						FirstCursor: "first",
						LastCursor:  "last",
						HasNext:     true,
						HasPrevious: true,
					},
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
				q := url.Values{}
				q.Add("room-id", "room-id")
				q.Add("user-id", "user-id")
				q.Add("cursor", "threepwood")
				q.Add("order", "asc")

				r, _ := http.NewRequest(http.MethodGet, "/api/v1/dice/rolls", nil)
				r.URL.RawQuery = q.Encode()
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
 ],
 "metadata": {
  "first_cursor": "first",
  "last_cursor": "last",
  "has_next": true,
  "has_previous": true
 }
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
			gotBody, err := io.ReadAll(res.Body)
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
			expBody:       "{\n \"Code\": 400,\n \"Message\": \"name is required\",\n \"Header\": null\n}",
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
			expBody:       "{\n \"Code\": 500,\n \"Message\": \"wanted error\",\n \"Header\": null\n}",
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
			gotBody, err := io.ReadAll(res.Body)
			require.NoError(err)
			assert.Equal(test.expStatusCode, res.StatusCode)
			assert.Equal(test.expBody, string(gotBody))
		})
	}
}

func TestAPIV1GetRoom(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")

	tests := map[string]struct {
		mock          func(*roommock.Service)
		req           func() *http.Request
		expStatusCode int
		expBody       string
	}{
		"Having an error while getting the room should fail.": {
			mock: func(m *roommock.Service) {
				m.On("GetRoom", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("wanted error"))
			},
			req: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/rooms/test-id", nil)
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusInternalServerError,
			expBody:       "{\n \"Code\": 500,\n \"Message\": \"wanted error\",\n \"Header\": null\n}",
		},

		"Having a correct request should get the room.": {
			mock: func(m *roommock.Service) {
				exp := room.GetRoomRequest{ID: "test-id"}
				resp := &room.GetRoomResponse{Room: model.Room{
					Name:      "test-room",
					CreatedAt: t0,
					ID:        "room-id",
				}}
				m.On("GetRoom", mock.Anything, exp).Once().Return(resp, nil)
			},
			req: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/rooms/test-id", nil)
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusOK,
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
			gotBody, err := io.ReadAll(res.Body)
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
			expBody:       "{\n \"Code\": 400,\n \"Message\": \"name is required\",\n \"Header\": null\n}",
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
			expBody:       "{\n \"Code\": 400,\n \"Message\": \"room_id is required\",\n \"Header\": null\n}",
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
			expBody:       "{\n \"Code\": 500,\n \"Message\": \"wanted error\",\n \"Header\": null\n}",
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
			gotBody, err := io.ReadAll(res.Body)
			require.NoError(err)
			assert.Equal(test.expStatusCode, res.StatusCode)
			assert.Equal(test.expBody, string(gotBody))
		})
	}
}

func TestAPIV1ListUsers(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")

	tests := map[string]struct {
		mock          func(*usermock.Service)
		req           func() *http.Request
		expStatusCode int
		expBody       string
	}{
		"Having a request without room id should fail.": {
			mock: func(m *usermock.Service) {},
			req: func() *http.Request {
				q := url.Values{}
				q.Add("room-id", "")
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
				r.URL.RawQuery = q.Encode()
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       "{\n \"Code\": 400,\n \"Message\": \"room-id is required\",\n \"Header\": null\n}",
		},

		"Having a request that fails while listing users, should fail.": {
			mock: func(m *usermock.Service) {
				m.On("ListUsers", mock.Anything, mock.Anything).Once().Return(nil, errors.New("wanted error"))
			},
			req: func() *http.Request {
				q := url.Values{}
				q.Add("room-id", "room-id")
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
				r.URL.RawQuery = q.Encode()
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusInternalServerError,
			expBody:       "{\n \"Code\": 500,\n \"Message\": \"wanted error\",\n \"Header\": null\n}",
		},

		"Having a request should list the users.": {
			mock: func(m *usermock.Service) {
				exp := user.ListUsersRequest{RoomID: "room-id"}
				resp := &user.ListUsersResponse{
					Users: []model.User{
						{
							ID:        "test1-id",
							RoomID:    "test1-id",
							Name:      "test1",
							CreatedAt: t0,
						},
						{
							ID:        "test2-id",
							RoomID:    "test2-id",
							Name:      "test2",
							CreatedAt: t0,
						},
					}}
				m.On("ListUsers", mock.Anything, exp).Once().Return(resp, nil)
			},
			req: func() *http.Request {
				q := url.Values{}
				q.Add("room-id", "room-id")
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
				r.URL.RawQuery = q.Encode()
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			expStatusCode: http.StatusOK,
			expBody: `{
 "items": [
  {
   "id": "test1-id",
   "name": "test1",
   "created_at": "1912-06-23T01:02:03Z"
  },
  {
   "id": "test2-id",
   "name": "test2",
   "created_at": "1912-06-23T01:02:03Z"
  }
 ]
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
			gotBody, err := io.ReadAll(res.Body)
			require.NoError(err)
			assert.Equal(test.expStatusCode, res.StatusCode)
			assert.Equal(test.expBody, string(gotBody))
		})
	}
}

func TestAPIV1WSRoomEvents(t *testing.T) {
	tests := map[string]struct {
		mock    func(*dicemock.Service)
		expBody string
		expErr  bool
	}{
		"Subscribing to dice roll created events in a room using websocket should subscribe and use the handler to send the events.": {
			mock: func(m *dicemock.Service) {
				// Expect subscription and send a dice roll created event in the moment the subscription is made.
				m.On("SubscribeDiceRollCreated", mock.Anything, mock.Anything).Once().Return(&dice.SubscribeDiceRollCreatedResponse{}, nil).Run(func(args mock.Arguments) {
					req := args[1].(dice.SubscribeDiceRollCreatedRequest)
					_ = req.EventHandler(context.TODO(), model.EventDiceRollCreated{
						DiceRoll: model.DiceRoll{},
					})
				})
			},
			expBody: "{\"metadata\":{\"type\":\"EventDiceRollCreated\"}}\n",
		},

		"Having an error while subscribing should return a websocket error.": {
			mock: func(m *dicemock.Service) {
				// Expect subscription and send a dice roll created event in the moment the subscription is made.
				m.On("SubscribeDiceRollCreated", mock.Anything, mock.Anything).Once().Return(&dice.SubscribeDiceRollCreatedResponse{}, errors.New("wanted error"))
			},
			expErr: true,
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

			// Prepare server and run, We use the server instead the regular handler because
			// we need a full websocket handshake.
			server := httptest.NewServer(h)
			defer server.Close()

			// Create a websocket connection.
			c, _, err := websocket.Dial(context.TODO(), server.URL+"/api/v1/ws/rooms/test-id", nil)
			require.NoError(err)
			defer c.Close(websocket.StatusNormalClosure, "")

			// Read a single message.
			_, gotBody, err := c.Read(context.TODO())

			// Check.
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expBody, string(gotBody))
			}
		})
	}
}
