package health

import (
	"fmt"
	"testing"
	"time"
)

type FakeChecker struct {
	ReturnArg0  interface{}
	ReturnArg1  error
	Invocations int
}

func (f *FakeChecker) Status() (interface{}, error) {
	f.Invocations++
	return f.ReturnArg0, f.ReturnArg1
}

func setupRunners(t *testing.T, cfgs []*Config) (*Health, []*Config, error) {
	h := New()
	testCheckInterval := time.Duration(10) * time.Millisecond

	if cfgs == nil {
		fakeChecker1 := &FakeChecker{}
		fakeChecker2 := &FakeChecker{}

		cfgs = []*Config{
			&Config{
				Name:     "foo",
				Checker:  fakeChecker1,
				Interval: testCheckInterval,
				Fatal:    false,
			},
			&Config{
				Name:     "bar",
				Checker:  fakeChecker2,
				Interval: testCheckInterval,
				Fatal:    false,
			},
		}
	}

	if err := h.AddChecks(cfgs); err != nil {
		return nil, nil, err
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
