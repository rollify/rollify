package ui

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/user"
)

const (
	formFieldManageUserUsername = "username"
	formFieldManageUserID       = "userID"
)

func (u ui) handlerActionManageUser() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse form info and return validation errors if any.
		username := r.FormValue(formFieldManageUserUsername)
		userID := r.FormValue(formFieldManageUserID)
		roomID := chi.URLParam(r, urlParamRoomID)

		// If we have a username then create a user.
		if username != "" {
			r, err := u.userAppSvc.CreateUser(r.Context(), user.CreateUserRequest{
				Name:   username,
				RoomID: roomID,
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

			userID = r.User.ID
		}

		if userID == "" {
			u.handleError(w, fmt.Errorf("user ID missing"))
			return
		}

		// TODO(slok): Check user ID exists.

		// Set user Id on cookie.
		cookies.SetUserID(w, roomID, userID, u.timeNow().Add(14*24*time.Hour))

		// Redirect to the room.
		u.redirectToURL(w, r, u.servePrefix+"/room/"+roomID)
	})
}
