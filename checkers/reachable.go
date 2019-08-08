package checkers

import (
	"net"
	"net/url"
	"time"
)

const (
	// ReachableDDHealthErrors is the datadog name used when there is a failure in the reachable checker
	ReachableDDHealthErrors = "health.errors"
	// ReachableDefaultPort is the default port used if no port is defined in a reachable checker
	ReachableDefaultPort = "80"
	// ReachableDefaultNetwork is the default network used in the reachable checker
	ReachableDefaultNetwork = "tcp"
)

var (
	// ReachableDefaultTimeout is the default timeout used when reachable is checking the URL
	ReachableDefaultTimeout = time.Duration(3) * time.Second
)

// ReachableDialer is the signature for a function that checks if an address is reachable
type ReachableDialer func(network, address string, timeout time.Duration) (net.Conn, error)

// ReachableDatadogIncrementer is any datadog client that has the Incr method for tracking metrics
type ReachableDatadogIncrementer interface {
	Incr(name string, tags []string, rate float64) error
}

// ReachableConfig is used for configuring an HTTP check. The only required field is `URL`.
//
// "Dialer" is optional and defaults to using net.DialTimeout.
//
// "Timeout" is optional and defaults to "3s".
//
// "Network" is optional and defaults to "tcp"; it should be one of "tcp",
// "tcp4", "tcp6", "unix", "unixpacket", "udp", "udp4", "udp6", "unixgram" or an
// IP transport. The IP transports are "ip", "ip4", or "ip6" followed by a colon
// and a literal protocol number or a protocol name, as in "ip:1" or "ip:icmp".
//
// "DatadogClient" is optional; if defined metrics will be sent via statsd.
//
// "DatadogTags" is optional; defines the tags that are passed to datadog when there is a failure
type ReachableConfig struct {
	URL           *url.URL                    // Required
	Dialer        ReachableDialer             // Optional (default net.DialTimeout)
	Timeout       time.Duration               // Optional (default 3s)
	Network       string                      // Optional (default tcp)
	DatadogClient ReachableDatadogIncrementer // Optional
	DatadogTags   []string                    // Optional
}

// ReachableChecker checks that URL responds to a TCP request
type ReachableChecker struct {
	dialer  ReachableDialer
	timeout time.Duration
	network string
	url     *url.URL
	datadog ReachableDatadogIncrementer
	tags    []string
}

// NewReachableChecker creates a new reachable health checker
func NewReachableChecker(cfg *ReachableConfig) (*ReachableChecker, error) {
	t := ReachableDefaultTimeout
	if cfg.Timeout != 0 {
		t = cfg.Timeout
	}
	d := net.DialTimeout
	if cfg.Dialer != nil {
		d = cfg.Dialer
	}
	n := ReachableDefaultNetwork
	if cfg.Network != "" {
		n = cfg.Network
	}
	r := &ReachableChecker{
		dialer:  d,
		timeout: t,
		network: n,
		url:     cfg.URL,
		datadog: cfg.DatadogClient,
		tags:    cfg.DatadogTags,
	}
	return r, nil
}

// Status checks if the endpoint is reachable
func (r *ReachableChecker) Status() (interface{}, error) {
	// We must provide a port so when a port is not set in the URL provided use
	// the default port (80)
	port := r.url.Port()
	if len(port) == 0 {
		port = ReachableDefaultPort
	}

	conn, err := r.dialer(r.network, r.url.Hostname()+":"+port, r.timeout)
	if err != nil {
		return r.fail(err)
	}
	if conn != nil {
		if errClose := conn.Close(); errClose != nil {
			return r.fail(errClose)
		}
	}
	return nil, nil
}

func (r *ReachableChecker) fail(err error) (interface{}, error) {
	if r.datadog != nil {
		err := r.datadog.Incr(ReachableDDHealthErrors, r.tags, 1.0)
		if err != nil{
			return nil, err
		}
	}
	return nil, err
}
