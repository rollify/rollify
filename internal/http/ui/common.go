package ui

import (
	"net/http"
)

func (u ui) tplRender(w http.ResponseWriter, templateName string, data any) {
	err := u.templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// tplDataCommon is a common data context that can be used on all templates and
// any template can use it.
type tplDataCommon struct {
	URLPrefix string
	Errors    []string
}

func (u ui) tplCommonData() tplDataCommon {
	return tplDataCommon{
		URLPrefix: u.servePrefix,
		Errors:    []string{},
	}
}
