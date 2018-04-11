package health

import (
	"fmt"
	"time"

	"github.com/InVisionApp/go-health/fakes"
	"github.com/InVisionApp/go-logger"
)

func setupRunners(cfgs []*Config, logger log.Logger) (*Health, []*Config, error) {
	h := New()
	testCheckInterval := time.Duration(10) * time.Millisecond

	if cfgs == nil {
		checker1 := &fakes.FakeICheckable{}
		checker2 := &fakes.FakeICheckable{}

		cfgs = []*Config{
			{
				Name:     "foo",
				Checker:  checker1,
				Interval: testCheckInterval,
				Fatal:    false,
			},
			{
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
	} else {
		h.Logger = log.NewNoop()
	}

	if err := h.Start(); err != nil {
		return nil, nil, err
	}

	// Correct number of runners/tickers were created
	if len(h.runners) != len(cfgs) {
		return nil, nil, fmt.Errorf("Start() did not create the expected number of runners")
	}

	return h, cfgs, nil
}
