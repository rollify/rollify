package apiv1

import (
	"errors"
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/rollify/rollify/internal/dice"
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

func (a *apiv1) listDiceTypes() restful.RouteFunction {
	logger := a.logger.WithKV(log.KV{"handler": "listDiceTypes"})

	return func(req *restful.Request, resp *restful.Response) {
		logger.Debugf("handler called")

		mr, err := a.diceAppSvc.ListDiceTypes(req.Request.Context())
		if err != nil {
			err := resp.WriteError(errToStatusCode(err), err)
			if err != nil {
				logger.Errorf("could not write http response: %w", err)
			}
			return
		}

		r := mapModelToAPIListDiceTypes(*mr)

		err = resp.WriteHeaderAndEntity(http.StatusOK, r)
		if err != nil {
			logger.Errorf("could not write http response: %w", err)
		}
	}
}

func errToStatusCode(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, dice.ErrNotValid):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
