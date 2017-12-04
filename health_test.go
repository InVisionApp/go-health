package health

import (
	"errors"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

var (
	testCheckInterval = time.Duration(10) * time.Millisecond
)

func TestNew(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should return a filled out instance of Health", func(t *testing.T) {
		h := New()

		Expect(h.configs).ToNot(BeNil())
		Expect(h.states).ToNot(BeNil())
		Expect(h.tickers).ToNot(BeNil())
	})
}

func TestAddChecks(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		h := New()
		testConfig := &Config{
			Name:     "foo",
			Checker:  &fakeChecker{},
			Interval: testCheckInterval,
			Fatal:    false,
		}

		err := h.AddChecks([]*Config{testConfig})

		Expect(err).To(BeNil())
		Expect(h.configs).To(ContainElement(testConfig))
		Expect(len(h.configs)).To(Equal(1))
	})

	t.Run("Should error if healthcheck is already running", func(t *testing.T) {
		h := New()
		h.active = true
		err := h.AddChecks([]*Config{})

		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(ErrNoAddCfgWhenActive))
	})

	t.Run("Should error if passed in empty config slice", func(t *testing.T) {
		h := New()
		err := h.AddChecks([]*Config{})

		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(ErrEmptyConfigs))
	})
}

func TestAddCheck(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		h := New()
		testConfig := &Config{
			Name:     "foo",
			Checker:  &fakeChecker{},
			Interval: testCheckInterval,
			Fatal:    false,
		}

		err := h.AddCheck(testConfig)

		Expect(err).To(BeNil())
		Expect(h.configs).To(ContainElement(testConfig))
		Expect(len(h.configs)).To(Equal(1))
	})

	t.Run("Should error if healthcheck is already running", func(t *testing.T) {
		h := New()
		h.active = true
		err := h.AddCheck(&Config{})

		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(ErrNoAddCfgWhenActive))
	})
}

func TestStart(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		h := New()
		checker1 := &fakeChecker{}
		checker2 := &fakeChecker{}

		cfgs := []*Config{
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

		err := h.AddChecks(cfgs)
		Expect(err).ToNot(HaveOccurred())

		err = h.Start()
		Expect(err).ToNot(HaveOccurred())
		// Correct number of runners/tickers were created
		Expect(len(h.tickers)).To(Equal(2))

		// Tickers are created (and saved) based on their name
		for _, v := range cfgs {
			Expect(h.tickers).To(HaveKey(v.Name))
		}

		// This is pretty brittle - will update if this is causing random test failures
		time.Sleep(time.Duration(15) * time.Millisecond)

		// Both runners should've ran
		Expect(checker1.Invocations).To(Equal(1), "Checker should have been executed")
		Expect(checker2.Invocations).To(Equal(1), "Checker should have been executed")

		// Both runners should've recorded their state
		Expect(h.states).To(HaveKey("foo"))
		Expect(h.states).To(HaveKey("bar"))
	})

	t.Run("Should error if healthcheck already running", func(t *testing.T) {
		h := New()

		err := h.AddCheck(&Config{})
		Expect(err).ToNot(HaveOccurred())

		// Set the healthcheck state to active
		h.active = true
		err = h.Start()
		Expect(err).To(Equal(ErrAlreadyRunning))
	})
}

func TestStop(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		h, _, err := setupRunners(t, nil)

		Expect(err).ToNot(HaveOccurred())
		Expect(h).ToNot(BeNil())

		err = h.Stop()
		Expect(err).ToNot(HaveOccurred())

		// Tickers map should be reset
		Expect(h.tickers).To(BeEmpty())
	})

	t.Run("Should error if healthcheck is not running", func(t *testing.T) {
		h := New()

		err := h.AddCheck(&Config{})
		Expect(err).ToNot(HaveOccurred())

		// Set the healthcheck state to active
		h.active = false
		err = h.Stop()
		Expect(err).To(Equal(ErrAlreadyStopped))
	})
}

func TestStartRunner(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path - checkers do not fail", func(t *testing.T) {
		h, cfgs, err := setupRunners(t, nil)

		Expect(err).ToNot(HaveOccurred())
		Expect(h).ToNot(BeNil())
		Expect(h.states).To(BeEmpty())

		// Brittle...
		time.Sleep(time.Duration(15) * time.Millisecond)

		// Did the ticker fire and create a state entry?
		for _, c := range cfgs {
			Expect(h.states).To(HaveKey(c.Name))
			Expect(h.states[c.Name].Status).To(Equal("ok"))
		}

		// Since nothing has failed, healthcheck should _not_ be in failed state
		Expect(h.failed).To(BeFalse())
	})

	t.Run("Happy path - 1 checker fails (non-fatal)", func(t *testing.T) {
		checker1 := &fakeChecker{}
		checker2 := &fakeChecker{}
		checker2.ReturnArg1 = errors.New("something failed")

		cfgs := []*Config{
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
		h, _, err := setupRunners(t, cfgs)

		Expect(err).ToNot(HaveOccurred())
		Expect(h).ToNot(BeNil())
		Expect(h.states).To(BeEmpty())

		// Brittle...
		time.Sleep(time.Duration(15) * time.Millisecond)

		// Did the ticker fire and create a state entry?
		for _, c := range cfgs {
			Expect(h.states).To(HaveKey(c.Name))
		}

		// First checker should've succeeded
		Expect(h.states[cfgs[0].Name].Status).To(Equal("ok"))

		// Second checker should've failed
		Expect(h.states[cfgs[1].Name].Status).To(Equal("failed"))
		Expect(h.states[cfgs[1].Name].Err).To(Equal(checker2.ReturnArg1))

		// Since nothing has failed, healthcheck should _not_ be in failed state
		Expect(h.failed).To(BeFalse())
	})

	t.Run("Happy path - 1 checker fails (fatal)", func(t *testing.T) {
		checker1 := &fakeChecker{}
		checker2 := &fakeChecker{}
		checker2.ReturnArg1 = errors.New("something failed")

		cfgs := []*Config{
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
				Fatal:    true,
			},
		}
		h, _, err := setupRunners(t, cfgs)

		Expect(err).ToNot(HaveOccurred())
		Expect(h).ToNot(BeNil())
		Expect(h.states).To(BeEmpty())

		// Brittle...
		time.Sleep(time.Duration(15) * time.Millisecond)

		// Did the ticker fire and create a state entry?
		for _, c := range cfgs {
			Expect(h.states).To(HaveKey(c.Name))
		}

		// First checker should've succeeded
		Expect(h.states[cfgs[0].Name].Status).To(Equal("ok"))

		// Second checker should've failed
		Expect(h.states[cfgs[1].Name].Status).To(Equal("failed"))
		Expect(h.states[cfgs[1].Name].Err).To(Equal(checker2.ReturnArg1))

		// Since second checker has failed fatally, global healthcheck state should be failed as well
		Expect(h.failed).To(BeTrue())
	})
}
