package ui

import (
	"net/http"
)

func (u ui) handlerFullIndex() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return the first page the user will see, it just renders the index page.
		d := struct {
			tplDataCommon
			tplDataCreateRoom
		}{
			tplDataCommon:     u.tplCommonData(),
			tplDataCreateRoom: tplDataCreateRoom{},
		}
		u.tplRender(w, "index", d)
	})
}
