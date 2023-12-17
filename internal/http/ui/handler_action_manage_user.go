package ui

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
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
		switch {
		case username != "":
			// Create user.
			us, err := u.userAppSvc.CreateUser(r.Context(), user.CreateUserRequest{
				Name:   username,
				RoomID: roomID,
			})
			if err != nil {
				u.handleError(w, fmt.Errorf("could not create user: %w", err))
				return
			}

			userID = us.User.ID

		case userID != "":
			// Check user exists.
			_, err := u.userAppSvc.GetUser(r.Context(), user.GetUserRequest{UserID: userID})
			if err != nil {
				u.handleError(w, fmt.Errorf("invalid user ID: %w", err))
				return
			}

		default:
			// Data missing, fail.
			u.handleError(w, fmt.Errorf("user ID or username missing"))
			return
		}

		// Set user Id on cookie.
		cookies.SetUserID(w, roomID, userID, u.timeNow().Add(14*24*time.Hour))

		// Redirect to the room.
		u.redirectToURL(w, r, u.servePrefix+"/room/"+roomID)
	})
}
