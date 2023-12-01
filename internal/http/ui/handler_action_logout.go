package ui

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (u ui) handlerActionLogout() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, urlParamRoomID)
		userID := cookies.GetUserID(r, roomID)

		roomRedirectURL := u.servePrefix + "/login/" + roomID

		if roomID == "" {
			u.redirectToURL(w, r, u.servePrefix)
		}

		if userID == "" {
			u.redirectToURL(w, r, roomRedirectURL)
			return
		}

		// Logout.
		cookies.DeleteUserID(w, roomID)

		// Redirect to the room login.
		u.redirectToURL(w, r, roomRedirectURL)
	})
}
