package apiv1

import (
	"errors"
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/rollify/rollify/internal/internalerrors"
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

		// Execute.
		mResp, err := a.diceAppSvc.ListDiceTypes(req.Request.Context())
		if err != nil {
			writeResponseError(logger, resp, errToStatusCode(err), err)
			logger.Warningf("error processing request: %s", err)
			return
		}

		// Map request.
		r := mapModelToAPIListDiceTypes(*mResp)
		err = resp.WriteHeaderAndEntity(http.StatusOK, r)
		if err != nil {
			logger.Errorf("could not write http response: %w", err)
		}
	}
}

func (a *apiv1) createDiceRoll() restful.RouteFunction {
	logger := a.logger.WithKV(log.KV{"handler": "createDiceRoll"})

	return func(req *restful.Request, resp *restful.Response) {
		logger.Debugf("handler called")

		// Map request.
		entReq := &createDiceRollRequest{}
		err := req.ReadEntity(entReq)
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}
		mReq, err := mapAPIToModelcreateDiceRoll(*entReq)
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}

		// Execute.
		mResp, err := a.diceAppSvc.CreateDiceRoll(req.Request.Context(), *mReq)
		if err != nil {
			writeResponseError(logger, resp, errToStatusCode(err), err)
			logger.Warningf("error processing request: %s", err)
			return
		}

		// Map response.
		r := mapModelToAPIcreateDiceRoll(*mResp)
		err = resp.WriteHeaderAndEntity(http.StatusCreated, r)
		if err != nil {
			logger.Errorf("could not write http response: %w", err)
		}
	}
}

func (a *apiv1) listDiceRolls() restful.RouteFunction {
	logger := a.logger.WithKV(log.KV{"handler": "listDiceRolls"})

	return func(req *restful.Request, resp *restful.Response) {
		logger.Debugf("handler called")

		// Map request.
		entReq := &listDiceRollsRequest{}
		err := req.ReadEntity(entReq)
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}
		mReq, err := mapAPIToModelListDiceRolls(*entReq)
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}

		// Execute.
		mResp, err := a.diceAppSvc.ListDiceRolls(req.Request.Context(), *mReq)
		if err != nil {
			writeResponseError(logger, resp, errToStatusCode(err), err)
			logger.Warningf("error processing request: %s", err)
			return
		}

		// Map response.
		r := mapModelToAPIListDiceRolls(*mResp)
		err = resp.WriteHeaderAndEntity(http.StatusOK, r)
		if err != nil {
			logger.Errorf("could not write http response: %w", err)
		}
	}
}

func (a *apiv1) createRoom() restful.RouteFunction {
	logger := a.logger.WithKV(log.KV{"handler": "createRoom"})

	return func(req *restful.Request, resp *restful.Response) {
		logger.Debugf("handler called")

		// Map request.
		entReq := &createRoomRequest{}
		err := req.ReadEntity(entReq)
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}
		mReq, err := mapAPIToModelCreateRoom(*entReq)
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}

		// Execute.
		mResp, err := a.roomAppSvc.CreateRoom(req.Request.Context(), *mReq)
		if err != nil {
			writeResponseError(logger, resp, errToStatusCode(err), err)
			logger.Warningf("error processing request: %s", err)
			return
		}

		// Map response.
		r := mapModelToAPICreateRoom(*mResp)
		err = resp.WriteHeaderAndEntity(http.StatusCreated, r)
		if err != nil {
			logger.Errorf("could not write http response: %w", err)
		}
	}
}

func (a *apiv1) createUser() restful.RouteFunction {
	logger := a.logger.WithKV(log.KV{"handler": "createUser"})

	return func(req *restful.Request, resp *restful.Response) {
		logger.Debugf("handler called")

		// Map request.
		entReq := &createUserRequest{}
		err := req.ReadEntity(entReq)
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}
		mReq, err := mapAPIToModelCreateUser(*entReq)
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}

		// Execute.
		mResp, err := a.UserAppSvc.CreateUser(req.Request.Context(), *mReq)
		if err != nil {
			writeResponseError(logger, resp, errToStatusCode(err), err)
			logger.Warningf("error processing request: %s", err)
			return
		}

		// Map response.
		r := mapModelToAPICreateUser(*mResp)
		err = resp.WriteHeaderAndEntity(http.StatusCreated, r)
		if err != nil {
			logger.Errorf("could not write http response: %w", err)
		}
	}
}

func (a *apiv1) listUsers() restful.RouteFunction {
	logger := a.logger.WithKV(log.KV{"handler": "listUsers"})

	return func(req *restful.Request, resp *restful.Response) {
		logger.Debugf("handler called")

		// Map request.
		entReq := &listUsersRequest{}
		err := req.ReadEntity(entReq)
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}
		mReq, err := mapAPIToModelListUsers(*entReq)
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}

		// Execute.
		mResp, err := a.UserAppSvc.ListUsers(req.Request.Context(), *mReq)
		if err != nil {
			writeResponseError(logger, resp, errToStatusCode(err), err)
			logger.Warningf("error processing request: %s", err)
			return
		}

		// Map response.
		r := mapModelToAPIListUsers(*mResp)
		err = resp.WriteHeaderAndEntity(http.StatusOK, r)
		if err != nil {
			logger.Errorf("could not write http response: %w", err)
		}
	}
}

func writeResponseError(logger log.Logger, resp *restful.Response, status int, err error) {
	err = resp.WriteServiceError(status, restful.NewError(status, err.Error()))
	if err != nil {
		logger.Errorf("could not write http response: %w", err)
	}
}

func errToStatusCode(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, internalerrors.ErrNotValid):
		return http.StatusNotFound
	case errors.Is(err, internalerrors.ErrAlreadyExists):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
