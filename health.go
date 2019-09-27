// Package health is a library that enables *async* dependency health checking for services running on an orchestrated container platform such as kubernetes or mesos.
//
// For additional overview, documentation and contribution guidelines, refer to the project's "README.md".
//
// For example usage, refer to https://module github.com/InVisionApp/go-health/v2/tree/master/examples/simple-http-server.
package health

import (
	"errors"
	"sync"
	"time"

	"github.com/InVisionApp/go-logger"
)

//go:generate counterfeiter -o ./fakes/icheckable.go . ICheckable

var (
	// ErrNoAddCfgWhenActive is returned when you attempt to add check(s) to an already active healthcheck instance
	ErrNoAddCfgWhenActive = errors.New("Unable to add new check configuration(s) while healthcheck is active")

	// ErrAlreadyRunning is returned when you attempt to "h.Start()" an already running healthcheck
	ErrAlreadyRunning = errors.New("Healthcheck is already running - nothing to start")

	// ErrAlreadyStopped is returned when you attempt to "h.Stop()" a non-running healthcheck instance
	ErrAlreadyStopped = errors.New("Healthcheck is not running - nothing to stop")

	// ErrEmptyConfigs is returned when you attempt to add an empty slice of configs via "h.AddChecks()"
	ErrEmptyConfigs = errors.New("Configs appears to be empty - nothing to add")
)

// The IHealth interface can be useful if you plan on replacing the actual health
// checker with a mock during testing. Otherwise, you can set "hc.Disable = true"
// after instantiation.
type IHealth interface {
	AddChecks(cfgs []*Config) error
	AddCheck(cfg *Config) error
	Start() error
	Stop() error
	State() (map[string]State, bool, error)
	Failed() bool
}

// ICheckable is an interface implemented by a number of bundled checkers such
// as "MySQLChecker", "RedisChecker" and "HTTPChecker". By implementing the
// interface, you can feed your own custom checkers into the health library.
type ICheckable interface {
	// Status allows you to return additional data as an "interface{}" and "error"
	// to signify that the check has failed. If "interface{}" is non-nil, it will
	// be exposed under "State.Details" for that particular check.
	Status() (interface{}, error)
}

// IStatusListener is an interface that handles health check failures and
// recoveries, primarily for stats recording purposes
type IStatusListener interface {
	// HealthCheckFailed is a function that handles the failure of a health
	// check event. This function is called when a health check state
	// transitions from passing to failing.
	// 	* entry - The recorded state of the health check that triggered the failure
	HealthCheckFailed(entry *State)

	// HealthCheckRecovered is a function that handles the recovery of a failed
	// health check.
	// 	* entry - The recorded state of the health check that triggered the recovery
	// 	* recordedFailures - the total failed health checks that lapsed
	// 	  between the failure and recovery
	//	* failureDurationSeconds - the lapsed time, in seconds, of the recovered failure
	HealthCheckRecovered(entry *State, recordedFailures int64, failureDurationSeconds float64)
}

// Config is a struct used for defining and configuring checks.
type Config struct {
	// Name of the check
	Name string

	// Checker instance used to perform health check
	Checker ICheckable

	// Interval between health checks
	Interval time.Duration

	// Fatal marks a failing health check so that the
	// entire health check request fails with a 500 error
	Fatal bool

	// Hook that gets called when this health check is complete
	OnComplete func(state *State)
}

// State is a struct that contains the results of the latest
// run of a particular check.
type State struct {
	// Name of the health check
	Name string `json:"name"`

	// Status of the health check state ("ok" or "failed")
	Status string `json:"status"`

	// Err is the error returned from a failed health check
	Err string `json:"error,omitempty"`

	// Fatal shows if the check will affect global result
	Fatal bool `json:"fatal,omitempty"`

	// Details contains more contextual detail about a
	// failing health check.
	Details interface{} `json:"details,omitempty"` // contains JSON message (that can be marshaled)

	// CheckTime is the time of the last health check
	CheckTime time.Time `json:"check_time"`

	ContiguousFailures int64     `json:"num_failures"`     // the number of failures that occurred in a row
	TimeOfFirstFailure time.Time `json:"first_failure_at"` // the time of the initial transitional failure for any given health check
}

// indicates state is failure
func (s *State) isFailure() bool {
	return s.Status == "failed"
}

// Health contains internal go-health internal structures.
type Health struct {
	Logger log.Logger

	// StatusListener will report failures and recoveries
	StatusListener IStatusListener

	active     *sBool // indicates whether the healthcheck is actively running
	configs    []*Config
	states     map[string]State
	statesLock sync.Mutex
	runners    map[string]chan struct{} // contains map of active runners w/ a stop channel
}

// New returns a new instance of the Health struct.
func New() *Health {
	return &Health{
		Logger:     log.NewSimple(),
		configs:    make([]*Config, 0),
		states:     make(map[string]State, 0),
		runners:    make(map[string]chan struct{}, 0),
		active:     newBool(),
		statesLock: sync.Mutex{},
	}
}

// DisableLogging will disable all logging by inserting the noop logger.
func (h *Health) DisableLogging() {
	h.Logger = log.NewNoop()
}

// AddChecks is used for adding multiple check definitions at once (as opposed
// to adding them sequentially via "AddCheck()").
func (h *Health) AddChecks(cfgs []*Config) error {
	if h.active.val() {
		return ErrNoAddCfgWhenActive
	}

	h.configs = append(h.configs, cfgs...)

	return nil
}

// AddCheck is used for adding a single check definition to the current health
// instance.
func (h *Health) AddCheck(cfg *Config) error {
	if h.active.val() {
		return ErrNoAddCfgWhenActive
	}

	h.configs = append(h.configs, cfg)
	return nil
}

// Start will start all of the defined health checks. Each of the checks run in
// their own goroutines (as "time.Ticker").
func (h *Health) Start() error {
	if h.active.val() {
		return ErrAlreadyRunning
	}

	// if there are no check configs, this is a noop
	if len(h.configs) < 1 {
		return nil
	}

	for _, c := range h.configs {
		h.Logger.WithFields(log.Fields{"name": c.Name}).Debug("Starting checker")
		ticker := time.NewTicker(c.Interval)
		stop := make(chan struct{})

		h.startRunner(c, ticker, stop)

		h.runners[c.Name] = stop
	}

	// Checkers are now actively running
	h.active.setTrue()

	return nil
}

// Stop will cause all of the running health checks to be stopped. Additionally,
// all existing check states will be reset.
func (h *Health) Stop() error {
	if !h.active.val() {
		return ErrAlreadyStopped
	}

	for name, stop := range h.runners {
		h.Logger.WithFields(log.Fields{"name": name}).Debug("Stopping checker")
		close(stop)
	}

	// Reset runner map
	h.runners = make(map[string]chan struct{}, 0)

	// Reset states
	h.safeResetStates()

	return nil
}

// State will return a map of all current healthcheck states (thread-safe), a
// bool indicating whether the healthcheck has fully failed and a potential error.
//
// The returned structs can be used for figuring out additional analytics or
// used for building your own status handler (as opposed to using the built-in
// "hc.HandlerBasic" or "hc.HandlerJSON").
//
// The map key is the name of the check.
func (h *Health) State() (map[string]State, bool, error) {
	return h.safeGetStates(), h.Failed(), nil
}

// Failed will return the basic state of overall health. This should be used when
// details about the failure are not needed
func (h *Health) Failed() bool {
	for _, val := range h.safeGetStates() {
		if val.Fatal && val.isFailure() {
			return true
		}
	}
	return false
}

func (h *Health) startRunner(cfg *Config, ticker *time.Ticker, stop <-chan struct{}) {

	// function to execute and collect check data
	checkFunc := func() {
		data, err := cfg.Checker.Status()

		stateEntry := &State{
			Name:      cfg.Name,
			Status:    "ok",
			Details:   data,
			CheckTime: time.Now(),
			Fatal:     cfg.Fatal,
		}

		if err != nil {
			h.Logger.WithFields(log.Fields{
				"check": cfg.Name,
				"fatal": cfg.Fatal,
				"err":   err,
			}).Error("healthcheck has failed")

			stateEntry.Err = err.Error()
			stateEntry.Status = "failed"
		}

		h.safeUpdateState(stateEntry)

		if cfg.OnComplete != nil {
			go cfg.OnComplete(stateEntry)
		}
	}

	go func() {
		defer ticker.Stop()

		// execute once so that it is immediate
		checkFunc()

		// all following executions
	RunLoop:
		for {
			select {
			case <-ticker.C:
				checkFunc()
			case <-stop:
				break RunLoop
			}
		}

		h.Logger.WithFields(log.Fields{"name": cfg.Name}).Debug("Checker exiting")
	}()
}

// resets the states in a concurrency-safe manner
func (h *Health) safeResetStates() {
	h.statesLock.Lock()
	defer h.statesLock.Unlock()
	h.states = make(map[string]State, 0)
}

// updates the check state in a concurrency-safe manner
func (h *Health) safeUpdateState(stateEntry *State) {
	// dispatch any status listeners
	h.handleStatusListener(stateEntry)

	// update states here
	h.statesLock.Lock()
	defer h.statesLock.Unlock()

	h.states[stateEntry.Name] = *stateEntry
}

// get all states in a concurrency-safe manner
func (h *Health) safeGetStates() map[string]State {
	h.statesLock.Lock()
	defer h.statesLock.Unlock()

	// deep copy h.states to avoid race
	statesCopy := make(map[string]State, 0)

	for k, v := range h.states {
		statesCopy[k] = v
	}

	return statesCopy
}

// if a status listener is attached
func (h *Health) handleStatusListener(stateEntry *State) {
	// get the previous state
	h.statesLock.Lock()
	prevState := h.states[stateEntry.Name]
	h.statesLock.Unlock()

	// state is failure
	if stateEntry.isFailure() {
		if !prevState.isFailure() {
			// new failure: previous state was ok
			if h.StatusListener != nil {
				go h.StatusListener.HealthCheckFailed(stateEntry)
			}

			stateEntry.TimeOfFirstFailure = time.Now()
		} else {
			// carry the time of first failure from the previous state
			stateEntry.TimeOfFirstFailure = prevState.TimeOfFirstFailure
		}
		stateEntry.ContiguousFailures = prevState.ContiguousFailures + 1
	} else if prevState.isFailure() {
		// recovery, previous state was failure
		failureSeconds := time.Now().Sub(prevState.TimeOfFirstFailure).Seconds()

		if h.StatusListener != nil {
			go h.StatusListener.HealthCheckRecovered(stateEntry, prevState.ContiguousFailures, failureSeconds)
		}
	}
}
