package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/InVisionApp/go-health"
)

type jsonStatus struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// NewBasicHandlerFunc will return an `http.HandlerFunc` that will write `ok`
// string + `http.StatusOK` to `rw`` if `h.Failed()` returns `false`;
// returns `error` + `http.StatusInternalServerError` if `h.Failed()` returns `true`.
func NewBasicHandlerFunc(h health.IHealth) http.HandlerFunc {
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

// NewJSONHandlerFunc will return an `http.HandlerFunc` that will marshal and
// write the contents of `h.StateMapInterface()` to `rw` and set status code to
//  `http.StatusOK` if `h.Failed()` is `false` OR set status code to
// `http.StatusInternalServerError` if `h.Failed` is `true`.
// It also accepts a set of optional custom fields to be added to the final JSON body
func NewJSONHandlerFunc(h health.IHealth, custom map[string]interface{}) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		states, failed, err := h.State()
		if err != nil {
			writeJSONStatus(rw, "error", fmt.Sprintf("Unable to fetch states: %v", err), http.StatusOK)
			return
		}

		msg := "ok"
		statusCode := http.StatusOK

		// There may be an _initial_ delay in display healthcheck data as the
		// healthchecks will only begin firing at "initialTime + checkIntervalTime"
		if len(states) == 0 {
			writeJSONStatus(rw, msg, "Healthcheck spinning up", statusCode)
			return
		}

		if failed {
			msg = "failed"
			statusCode = http.StatusInternalServerError
		}

		fullBody := map[string]interface{}{
			"status":  msg,
			"details": states,
		}

		for k, v := range custom {
			if k != "status" && k != "details" {
				fullBody[k] = v
			}
		}

		data, err := json.Marshal(fullBody)
		if err != nil {
			writeJSONStatus(rw, "error", fmt.Sprintf("Failed to marshal state data: %v", err), http.StatusOK)
			return
		}

		writeJSONResponse(rw, statusCode, data)
	})
}

func writeJSONStatus(rw http.ResponseWriter, status, message string, statusCode int) {
	jsonData, _ := json.Marshal(&jsonStatus{
		Message: message,
		Status:  status,
	})

	writeJSONResponse(rw, statusCode, jsonData)
}

func writeJSONResponse(rw http.ResponseWriter, statusCode int, content []byte) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	rw.Write(content)
}
