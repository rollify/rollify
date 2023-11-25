package ui

import (
	"net/http"
)

func (u ui) handlerFullIndex() http.HandlerFunc {
	type tplData struct {
		tplDataCommon
		tplDataCreateRoom
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u.tplRender(w, "index", tplData{
			tplDataCommon:     u.tplCommonData(),
			tplDataCreateRoom: tplDataCreateRoom{},
		})
	})
}
