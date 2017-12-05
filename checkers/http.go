package checkers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout = time.Duration(3) * time.Second
)

// HTTPConfig is used for configuring an HTTP check. The only required field is `URL`.
//
// - `Method` is optional and defaults to `GET` if undefined
// - `Payload` is optional and can accept `string`, `[]byte` or will attempt to
// marshal the input to JSON for use w/ `bytes.NewReader()`
// - `StatusCode` is optional and defaults to `200`
// - `Expect` is optional; if defined, operates as a basic "body should contain <string>"
// - `Client` is optional; if undefined, a new client will be created using `Timeout`
// - `Timeout` is optional and defaults to `3s`
type HTTPConfig struct {
	URL        *url.URL      // Required
	Method     string        // Optional (default GET)
	Payload    interface{}   // Optional
	StatusCode int           // Optional (default 200)
	Expect     string        // Optional
	Client     *http.Client  // Optional
	Timeout    time.Duration // Optional (default 3s)
}

// HTTP implements the ICheckable interface
type HTTP struct {
	Config *HTTPConfig
}

// NewHTTP creates a new HTTP checker that can be used for `.AddCheck(s)`.
func NewHTTP(cfg *HTTPConfig) (*HTTP, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Passed in config cannot be nil")
	}

	if err := cfg.prepare(); err != nil {
		return nil, fmt.Errorf("Unable to prepare given config: %v", err)
	}

	return &HTTP{
		Config: cfg,
	}, nil
}

// Status is used for performing an HTTP check against a dependency; it satisfies
// the `ICheckable` interface.
func (h *HTTP) Status() error {
	resp, err := h.do()
	if err != nil {
		return err
	}

	// Check if StatusCode matches
	if resp.StatusCode != h.Config.StatusCode {
		return fmt.Errorf("Received status code '%v' does not match expected status code '%v'",
			resp.StatusCode, h.Config.StatusCode)
	}

	// If Expect is set, verify if returned response contains expected data
	if h.Config.Expect != "" {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Unable to read response body to perform content expectancy check: %v", err)
		}
		defer resp.Body.Close()

		if !strings.Contains(string(data), h.Config.Expect) {
			return fmt.Errorf("Received response body '%v' does not contain expected content '%v'",
				string(data), h.Config.Expect)
		}
	}

	return nil
}

func (h *HTTP) do() (*http.Response, error) {
	payload, err := parsePayload(h.Config.Payload)
	if err != nil {
		return nil, fmt.Errorf("error parsing payload: %v", err)
	}

	req, err := http.NewRequest(h.Config.Method, h.Config.URL.String(), payload)
	if err != nil {
		return nil, fmt.Errorf("Unable to create new HTTP request for HTTPMonitor check: %v", err)
	}

	resp, err := h.Config.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ran into error while performing '%v' request: %v", h.Config.Method, err)
	}

	return resp, nil
}

func (h *HTTPConfig) prepare() error {
	if h.URL == nil {
		return errors.New("URL cannot be nil")
	}

	// Default StatusCode to 200
	if h.StatusCode == 0 {
		h.StatusCode = http.StatusOK
	}

	// Default to GET
	if h.Method == "" {
		h.Method = "GET"
	}

	if h.Timeout == 0 {
		h.Timeout = defaultHTTPTimeout
	}

	if h.Client == nil {
		h.Client = &http.Client{Timeout: h.Timeout}
	} else {
		h.Client.Timeout = h.Timeout
	}

	return nil
}

func parsePayload(b interface{}) (io.Reader, error) {
	if b == nil {
		return nil, nil
	}

	switch b.(type) {
	case []byte:
		return bytes.NewReader(b.([]byte)), nil
	case string:
		return bytes.NewReader([]byte(b.(string))), nil
	default:
		jb, err := json.Marshal(b)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json body: %v", err)
		}

		return bytes.NewReader(jb), nil
	}
}
