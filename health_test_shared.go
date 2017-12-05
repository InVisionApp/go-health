package health

import (
	"fmt"
	"time"

	"github.com/InVisionApp/go-health/fakes"
	"github.com/InVisionApp/go-health/log"
)

func setupRunners(cfgs []*Config, logger log.ILogger) (*Health, []*Config, error) {
	h := New()
	testCheckInterval := time.Duration(10) * time.Millisecond

	if cfgs == nil {
		checker1 := &fakes.FakeICheckable{}
		checker2 := &fakes.FakeICheckable{}

		cfgs = []*Config{
			&Config{
				Name:     "foo",
				Checker:  checker1,
				Interval: testCheckInterval,
				Fatal:    false,
			},
			&Config{
				Name:     "bar",
				Checker:  checker2,
				Interval: testCheckInterval,
				Fatal:    false,
			},
		}
	}

	if err := h.AddChecks(cfgs); err != nil {
		return nil, nil, err
	}

	if logger != nil {
		h.Logger = logger
	}

	if err := h.Start(); err != nil {
		return nil, nil, err
	}

	// Correct number of runners/tickers were created
	if len(h.tickers) != len(cfgs) {
		return nil, nil, fmt.Errorf("Start() did not create the expected number of tickers")
	}

	return h, cfgs, nil
}
