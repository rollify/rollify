package htmx

import (
	"net/http"
	"strconv"
)

// https://htmx.org/reference/#Request_headers
type Request struct {
	headers http.Header
}

func NewRequest(h http.Header) *Request {
	return &Request{headers: h}
}

func (r *Request) IsBoosted() bool {
	is := r.headers.Get("HX-Boosted")
	return r.parseStrBool(is)
}

func (r *Request) CurrentURL() string {
	return r.headers.Get("HX-Current-URL")
}

func (r *Request) IsHistoryRestoreRequest() bool {
	is := r.headers.Get("HX-History-Restore-Request")
	return r.parseStrBool(is)
}

func (r *Request) Prompt() string {
	return r.headers.Get("HX-Prompt")
}

func (r *Request) TargetID() string {
	return r.headers.Get("HX-Target")
}

func (r *Request) TriggerName() string {
	return r.headers.Get("HX-Trigger-Name")
}

func (r *Request) TriggerID() string {
	return r.headers.Get("HX-Trigger")
}

func (r *Request) IsHTMXRequest() bool {
	is := r.headers.Get("HX-Request")
	return r.parseStrBool(is)
}

func (r *Request) parseStrBool(s string) bool {
	result, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return result
}
