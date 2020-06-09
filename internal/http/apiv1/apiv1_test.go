package apiv1_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/http/apiv1"
)

func TestAPIV1Pong(t *testing.T) {
	tests := map[string]struct {
		req           func() *http.Request
		expStatusCode int
		expBody       []byte
	}{
		"Having a correct request for ping, should be handled correctly.": {
			req: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "/api/v1/ping", nil)
				return r
			},
			expStatusCode: http.StatusOK,
			expBody:       []byte(`"pong"`),
		},

		"Having a incorrect method request for ping, should not be found.": {
			req: func() *http.Request {
				r, _ := http.NewRequest(http.MethodPost, "/api/v1/ping", nil)
				return r
			},
			expStatusCode: http.StatusMethodNotAllowed,
			expBody:       []byte(`405: Method Not Allowed`),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Prepare.
			cfg := apiv1.Config{}
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
			assert.Equal(test.expBody, gotBody)
		})
	}
}
