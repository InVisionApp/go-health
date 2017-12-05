package health

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// NewBasicHandler will return an `http.HandlerFunc` that will write `ok` string + `http.StatusOK` to `rw`` if `h.Failed()`
// returns `false`; returns `error` + `http.StatusInternalServerError` if
// `h.Failed()` returns `true`.
func NewBasicHandlerFunc(h IHealth) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		body := "ok"

		if h.Failed() {
			status = http.StatusInternalServerError
			body = "failed"
		}

		rw.WriteHeader(status)
		rw.Write([]byte(body))
	})
}

// NewJSONHandler will return an `http.HandlerFunc` that will marshal and write the contents of `h.StateMapInterface()` to
// `rw` and set status code to `http.StatusOK` if `h.Failed()` is `false` OR
// set status code to `http.StatusInternalServerError` if `h.Failed` is `true`.
func NewJSONHandler(h IHealth) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		msg := "ok"
		state, failed, _ := h.State()
		if failed {
			status = http.StatusInternalServerError
			msg = "failed"
		}

		fullBody := map[string]interface{}{
			"status":  msg,
			"details": state,
		}

		stateJSON, err := json.Marshal(fullBody)
		if err != nil {
			stateJSON = []byte(fmt.Sprintf(
				`{
					"status": "error",
					"details": "failed to marshal state details: %v"
				}`, err))
		}

		rw.WriteHeader(status)
		rw.Write(stateJSON)
	})
}
