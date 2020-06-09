package apiv1

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/rollify/rollify/internal/log"
)

func (a *apiv1) pong() restful.RouteFunction {
	logger := a.logger.WithKV(log.KV{"handler": "pong"})

	return func(req *restful.Request, resp *restful.Response) {
		logger.Debugf("handler called")

		err := resp.WriteHeaderAndEntity(http.StatusOK, "pong")
		if err != nil {
			logger.Errorf("could not write http response: %w", err)
		}
	}
}
