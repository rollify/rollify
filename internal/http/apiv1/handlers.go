package apiv1

import (
	"context"
	"errors"
	"net/http"

	"github.com/emicklei/go-restful"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/model"
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
		mReq, err := mapAPIToModelListDiceRolls(req.Request.URL.Query())
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

func (a *apiv1) getRoom() restful.RouteFunction {
	logger := a.logger.WithKV(log.KV{"handler": "getRoom"})

	return func(req *restful.Request, resp *restful.Response) {
		logger.Debugf("handler called")

		// Map request.
		mReq, err := mapAPIToModelGetRoom(req.PathParameters())
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}

		// Execute.
		mResp, err := a.roomAppSvc.GetRoom(req.Request.Context(), *mReq)
		if err != nil {
			writeResponseError(logger, resp, errToStatusCode(err), err)
			logger.Warningf("error processing request: %s", err)
			return
		}

		// Map response.
		r := mapModelToAPIGetRoom(*mResp)
		err = resp.WriteHeaderAndEntity(http.StatusOK, r)
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
		mResp, err := a.userAppSvc.CreateUser(req.Request.Context(), *mReq)
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
		mReq, err := mapAPIToModelListUsers(req.Request.URL.Query())
		if err != nil {
			writeResponseError(logger, resp, http.StatusBadRequest, err)
			return
		}

		// Execute.
		mResp, err := a.userAppSvc.ListUsers(req.Request.Context(), *mReq)
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

func (a *apiv1) wsRoomEvents() restful.RouteFunction {
	const wsRoomEventsRoomID = "id"

	logger := a.logger.WithKV(log.KV{"handler": "wsRoomEvents"})

	return func(req *restful.Request, resp *restful.Response) {
		logger.Debugf("handler called")

		// Get correct data.
		roomID := req.PathParameters()[wsRoomEventsRoomID]

		// Upgrade connection to websocket.
		c, err := websocket.Accept(resp.ResponseWriter, req.Request, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			writeResponseError(logger, resp, errToStatusCode(err), err)
			logger.Warningf("error processing websocket request: %s", err)
			return
		}
		defer c.Close(websocket.StatusInternalError, "")
		logger.Debugf("websocket connected")
		defer logger.Debugf("websocket disconnected")

		// Subscribe user to room events.
		modelReq := dice.SubscribeDiceRollCreatedRequest{
			RoomID: roomID,
			EventHandler: func(ctx context.Context, e model.EventDiceRollCreated) error {
				resp := mapModelToAPIWSDiceRollCreatedEvent(e)
				return wsjson.Write(ctx, c, resp)
			},
		}
		modelResp, err := a.diceAppSvc.SubscribeDiceRollCreated(req.Request.Context(), modelReq)
		if err != nil {
			logger.Warningf("error subscribing websocket to dice roll created events: %s", err)
			return
		}
		defer func() {
			err := modelResp.UnsubscribeFunc()
			if err != nil {
				logger.Warningf("error subscribing websocket to dice roll created events: %s", err)
			}
		}()

		// We don't plan to receive any message from the websocket, only send,
		// that's why we use `CloseRead` and wait until we are done.
		ctx := c.CloseRead(req.Request.Context())
		<-ctx.Done()
		c.Close(websocket.StatusNormalClosure, "")
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
		return http.StatusBadRequest
	case errors.Is(err, internalerrors.ErrMissing):
		return http.StatusNotFound
	case errors.Is(err, internalerrors.ErrAlreadyExists):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
