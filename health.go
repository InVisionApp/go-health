// A library that enables async dependency health checking for services running on an orchastrated container 
// platform such as kubernetes or mesos.
package health

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/InVisionApp/go-health/loggers"
)

//go:generate counterfeiter -o ./fakes/icheckable.go . ICheckable

var (
	// ErrNoAddCfgWhenActive is returned when you attempt to add check(s) to an already active healthcheck instance
	ErrNoAddCfgWhenActive = errors.New("Unable to add new check configuration(s) while healthcheck is active")

	// ErrAlreadyRunning is returned when you attempt to `h.Start()` an already running healthcheck
	ErrAlreadyRunning = errors.New("Healthcheck is already running - nothing to start")

	// ErrAlreadyStopped is returned when you attempt to `h.Stop()` a non-running healthcheck instance
	ErrAlreadyStopped = errors.New("Healthcheck is not running - nothing to stop")

	// ErrEmptyConfigs is returned when you attempt to add an empty slice of configs via `h.AddChecks()`
	ErrEmptyConfigs = errors.New("Configs appears to be empty - nothing to add")
)

// The IHealth interface can be useful if you plan on replacing the actual health
// checker with a mock during testing. Otherwise, you can set `hc.Disable = true`
// after instantiation.
type IHealth interface {
	AddChecks(cfgs []*Config) error
	AddCheck(cfg *Config) error
	Start() error
	Stop() error
	State() (map[string]State, bool, error)
	Failed() bool
}

// The ICheckable interface is implemented by a number of bundled checkers such
// as `MySQLChecker`, `RedisChecker` and `HTTPChecker`. By implementing the
// interface, you can feed your own custom checkers into the health library.
type ICheckable interface {
	// Status allows you to return additional data as an `interface{}` and `error`
	// to signify that the check has failed. If `interface{}` is non-nil, it will
	// be exposed under `State.Details` for that particular check.
	Status() (interface{}, error)
}

// The Config struct is used for defining and configuring checks.
type Config struct {
	Name     string
	Checker  ICheckable
	Interval time.Duration
	Fatal    bool
}

// The State struct contains the results of the latest run of a particular check.
type State struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Err    string `json:"error,omitempty"`
	// contains JSON message (that can be marshaled)
	Details   interface{} `json:"details,omitempty"`
	CheckTime time.Time   `json:"check_time"`
}

// Health contains internal go-health internal structures
type Health struct {
	Logger loggers.ILogger

	active *sBool // indicates whether the healthcheck is actively running
	failed *sBool // indicates whether the healthcheck has encountered a fatal error in one of its deps

	configs    []*Config
	states     map[string]State
	statesLock sync.Mutex
	tickers    map[string]*time.Ticker // contains map of actively running tickers
}

// New returns a new instance of the Health struct.
func New() *Health {
	return &Health{
		Logger:     loggers.NewBasic(),
		configs:    make([]*Config, 0),
		states:     make(map[string]State, 0),
		tickers:    make(map[string]*time.Ticker, 0),
		active:     newBool(),
		failed:     newBool(), // init as false
		statesLock: sync.Mutex{},
	}
}

// DisableLogging will disable all logging by inserting the noop logger
func (h *Health) DisableLogging() {
	h.Logger = loggers.NewNoop()
}

// AddChecks is used for adding multiple check definitions at once (as opposed
// to adding them sequentially via `AddCheck()`).
func (h *Health) AddChecks(cfgs []*Config) error {
	if h.active.val() {
		return ErrNoAddCfgWhenActive
	}

	if len(cfgs) == 0 {
		return ErrEmptyConfigs
	}

	h.configs = cfgs
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
// their own goroutines (as `time.Ticker`).
func (h *Health) Start() error {
	if h.active.val() {
		return ErrAlreadyRunning
	}

	for _, c := range h.configs {
		h.Logger.Debug("Starting checker", map[string]interface{}{"name": c.Name})
		ticker := time.NewTicker(c.Interval)

		if err := h.startRunner(c, ticker); err != nil {
			return fmt.Errorf("Unable to create healthcheck runner '%v': %v", c.Name, err)
		}

		h.tickers[c.Name] = ticker
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

	for name, ticker := range h.tickers {
		h.Logger.Debug("Stopping checker", map[string]interface{}{"name": name})
		ticker.Stop()
	}

	// Reset ticker map
	h.tickers = make(map[string]*time.Ticker, 0)

	return nil
}

// State will return a map of all current healthcheck states (thread-safe), a
// bool indicating whether the healthcheck has fully failed and a potential error.
//
// The returned structs can be used for figuring out additional analytics or
// used for building your own status handler (as opposed to using the built-in
// `hc.HandlerBasic` or `hc.HandlerJSON`).
//
// The map key is the name of the check.
func (h *Health) State() (map[string]State, bool, error) {
	return h.safeGetStates(), h.failed.val(), nil
}

// Failed will return the basic state of overall health. This should be used when
// details about the failure are not needed
func (h *Health) Failed() bool {
	return h.failed.val()
}

func (h *Health) startRunner(cfg *Config, ticker *time.Ticker) error {

	// function to execute and collect check data
	checkFunc := func() {
		data, err := cfg.Checker.Status()

		stateEntry := &State{
			Name:      cfg.Name,
			Status:    "ok",
			Details:   data,
			CheckTime: time.Now(),
		}

		if err != nil {
			h.Logger.Error("healthcheck has failed", map[string]interface{}{
				"check": cfg.Name,
				"fatal": cfg.Fatal,
				"err":   err,
			})

			stateEntry.Err = err.Error()
			stateEntry.Status = "failed"
		}

		// Toggle the global state failure if this check is allowed to cause
		// a complete healthcheck failure.
		if err != nil && cfg.Fatal {
			h.failed.setTrue()
		}

		h.safeUpdateState(stateEntry)
	}

	go func() {
		// execute once so that it is immediate
		checkFunc()

		// all following executions
		for range ticker.C {
			checkFunc()
		}

		h.Logger.Debug("Checker exiting", map[string]interface{}{"name": cfg.Name})
	}()

	return nil
}

func (h *Health) safeUpdateState(stateEntry *State) {
	// update states here
	h.statesLock.Lock()
	defer h.statesLock.Unlock()
	h.states[stateEntry.Name] = *stateEntry
}

func (h *Health) safeGetStates() map[string]State {
	h.statesLock.Lock()
	defer h.statesLock.Unlock()
	return h.states
}
