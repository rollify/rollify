package apiv1_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
