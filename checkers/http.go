package checkers

import (
	"fmt"
	"net/http"
	"net/url"
)

type HTTPConfig struct {
	URL        *url.URL
	Method     string
	Payload    interface{}
	StatusCode int
	Expect     string
	Client     *http.Client
}

type HTTP struct {
	Config *HTTPConfig
}

func NewHTTP(cfg *HTTPConfig) (*HTTP, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Passed in config cannot be empty")
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("Unable to validate given config: %v", err)
	}

	return &HTTP{
		Config: cfg,
	}, nil
}

func (h *HTTPConfig) validate() error {
	return nil
}

func (h *HTTP) Status() error {
	panic("implement me")
}
