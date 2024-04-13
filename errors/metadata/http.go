package metadata

import nethttp "net/http"

type http struct {
	code       int
	statusText string
}

func (h *http) setHTTPError(code int) *http {
	h.code = code
	h.statusText = nethttp.StatusText(code)

	return h
}

func (h http) getHTTPCode() int {
	return h.code
}

func (h http) getHTTPStatus() string {
	return h.statusText
}
