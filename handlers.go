package health

import (
	"net/http"
)

// BasicHandler will write `ok` string + `http.StatusOK` to `rw`` if `h.Failed()`
// returns `false`; returns `error` + `http.StatusInternalServerError` if
// `h.Failed()` returns `true`.
func (h *Health) BasicHandler(rw http.ResponseWriter, r *http.Request) {
}

// JSONHandler will marshal and write the contents of `h.StateMapInterface()` to
// `rw` and set status code to `http.StatusOK` if `h.Failed()` is `false` OR
// set status code to `http.StatusInternalServerError` if `h.Failed` is `true`.
func (h *Health) JSONHandler(rw http.ResponseWriter, r *http.Request) {
}
