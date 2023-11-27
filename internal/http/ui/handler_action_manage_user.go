package ui

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/user"
)

const (
	formFieldManageUserUsername = "username"
	formFieldManageUserID       = "userID"
)

type reqManageUser struct {
	Username string
	UserID   string
	RoomID   string
}

func uiToModelManageUser(r *http.Request) *reqManageUser {
	username := r.FormValue(formFieldManageUserUsername)
	userID := r.FormValue(formFieldManageUserID)
	roomID := chi.URLParam(r, urlParamRoomID)

	return &reqManageUser{
		Username: strings.TrimSpace(username),
		UserID:   strings.TrimSpace(userID),
		RoomID:   strings.TrimSpace(roomID),
	}
}

func (u ui) handlerActionManageUser() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse form info and return validation errors if any.
		m := uiToModelManageUser(r)

		// If we have a username then create a user.
		if m.Username != "" {
			r, err := u.userAppSvc.CreateUser(r.Context(), user.CreateUserRequest{
				Name:   m.Username,
				RoomID: m.RoomID,
			})
			if err != nil {
				if !errors.Is(err, internalerrors.ErrAlreadyExists) {
					u.handleError(w, fmt.Errorf("could not create user: %w", err))
					return
				} else {
					// TODO(slok): Remove this error and Get user model (with ID) using the username.
					u.handleError(w, fmt.Errorf("Not implemented"))
				}
			}

			m.UserID = r.User.ID
		}

		if m.UserID == "" {
			u.handleError(w, fmt.Errorf("user ID missing"))
			return
		}

		// TODO(slok): Check user ID exists.

		// Set user Id on cookie.
		cookies.SetUserID(w, m.RoomID, m.UserID, u.timeNow().Add(14*24*time.Hour))

		// Redirect to the room.
		u.redirectToURL(w, r, u.servePrefix+"/room/"+m.RoomID)
	})
}
