package htmx

import "net/http"

// https://htmx.org/reference/#response_headers
type Response struct {
	headers map[string]string
}

func NewResponse() *Response {
	return &Response{headers: map[string]string{}}
}

func (r *Response) WithLocation(url string) *Response {
	r.headers["HX-Location"] = url
	return r
}

func (r *Response) WithPushURL(url string) *Response {
	r.headers["HX-Push-Url"] = url
	return r
}

func (r *Response) WithRedirect(url string) *Response {
	r.headers["HX-Redirect"] = url
	return r
}

func (r *Response) WithRefresh() *Response {
	r.headers["HX-Refresh"] = "true"
	return r
}

func (r *Response) WithReplaceURL(url string) *Response {
	r.headers["HX-Replace-Url"] = url
	return r
}

type SwapMode string

const (
	SwapModeInnerHTML   SwapMode = "innerHTML"
	SwapModeOuterHTML   SwapMode = "outerHTML"
	SwapModeBeforebegin SwapMode = "beforebegin"
	SwapModeAfterbegin  SwapMode = "afterbegin"
	SwapModeBeforeend   SwapMode = "beforeend"
	SwapModeAfterend    SwapMode = "afterend"
	SwapModeDelete      SwapMode = "delete"
	SwapModeNone        SwapMode = "none"
)

func (r *Response) WithReswap(m SwapMode) *Response {
	r.headers["HX-Reswap"] = string(m)
	return r
}

func (r *Response) WithRetarget(cssSelector string) *Response {
	r.headers["HX-Retarget"] = cssSelector
	return r
}

func (r *Response) WithReselect(cssSelector string) *Response {
	r.headers["HX-Reselect"] = cssSelector
	return r
}

func (r *Response) WithTrigger(event string) *Response {
	r.headers["HX-Trigger"] = event
	return r
}

func (r *Response) WithTriggerAfterSettle(event string) *Response {
	r.headers["HX-Trigger-After-Settle"] = event
	return r
}

func (r *Response) WithTriggerAfterSwap(event string) *Response {
	r.headers["HX-Trigger-After-Swap"] = event
	return r
}

func (r *Response) SetHeaders(w http.ResponseWriter) {
	for k, v := range r.headers {
		w.Header().Set(k, v)
	}
}
