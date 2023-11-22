package ui

import (
	"fmt"
	"net/http"
	"time"
)

const cookieUserID = "_room_user_id_%s"

// cookies knows how to handle cookies for our application.
var cookies = cookieManager{}

type cookieManager struct{}

func (c cookieManager) SetUserID(w http.ResponseWriter, roomID, userID string, expiration time.Time) {
	c.set(w, fmt.Sprintf(cookieUserID, roomID), userID, expiration)
}

func (c cookieManager) GetUserID(r *http.Request, roomID string) string {
	return c.get(r, fmt.Sprintf(cookieUserID, roomID))
}

func (c cookieManager) DeleteUserID(w http.ResponseWriter, roomID string) {
	c.delete(w, fmt.Sprintf(cookieUserID, roomID))
}

func (c cookieManager) set(w http.ResponseWriter, k, v string, expiration time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:    k,
		Value:   v,
		Path:    "/",
		Expires: expiration,
	})
}

func (c cookieManager) get(r *http.Request, k string) string {
	ck, err := r.Cookie(k)
	if err != nil {
		return ""
	}

	return ck.Value
}

func (c cookieManager) delete(w http.ResponseWriter, k string) {
	http.SetCookie(w, &http.Cookie{
		Name:    k,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})
}
