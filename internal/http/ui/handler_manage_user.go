package ui

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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

func uiToModelManageUser(r *http.Request) (m *reqManageUser, errMsgs []string) {
	username := r.FormValue(formFieldManageUserUsername)
	userID := r.FormValue(formFieldManageUserID)
	roomID := chi.URLParam(r, urlParamRoomID)

	return &reqManageUser{
		Username: strings.TrimSpace(username),
		UserID:   strings.TrimSpace(userID),
		RoomID:   strings.TrimSpace(roomID),
	}, nil
}

func (u ui) manageUser() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse form info and return validation errors if any.
		m, errs := uiToModelManageUser(r)
		if errs != nil {
			d := tplDataCreateRoom{
				tplDataCommon: tplDataCommon{URLPrefix: u.servePrefix},
				FormErrors:    errs,
			}
			u.tplRender(w, "create_room_form", d)
			return
		}

		// If we have a username then create a user.
		if m.Username != "" {
			r, err := u.userAppSvc.CreateUser(r.Context(), user.CreateUserRequest{
				Name:   m.Username,
				RoomID: m.RoomID,
			})
			if err != nil {
				u.handleError(w, fmt.Errorf("could not create room: %w", err))
				return
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
