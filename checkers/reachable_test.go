package checkers_test

import (
	"context"
	"errors"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/InVisionApp/go-health/v2/checkers"
	"github.com/InVisionApp/go-health/v2/fakes"
	"github.com/InVisionApp/go-health/v2/fakes/netfakes"
)

func TestReachableSuccessUsingDefaults(t *testing.T) {
	assert := assert.New(t)
	dd := &fakes.FakeReachableDatadogIncrementer{}
	u, _ := url.Parse("http://example.com")
	cfg := &checkers.ReachableConfig{
		URL: u,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			assert.Equal(checkers.ReachableDefaultNetwork, network)
			assert.Equal(u.Hostname()+":"+checkers.ReachableDefaultPort, address)
			assert.Equal(checkers.ReachableDefaultTimeout, timeout)
			return nil, nil
		},
		DatadogClient: dd,
		DatadogTags: []string{
			"dependency:test-service",
		},
	}
	c, err := checkers.NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status(context.TODO())
	assert.NoError(err)
	assert.Equal(0, dd.IncrCallCount())
}

func TestReachableSuccess(t *testing.T) {
	assert := assert.New(t)
	dd := &fakes.FakeReachableDatadogIncrementer{}
	u, _ := url.Parse("http://example.com:8080")
	n := "udp"
	to := time.Duration(10) * time.Second
	cfg := &checkers.ReachableConfig{
		URL:     u,
		Network: n,
		Timeout: to,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			assert.Equal(n, network)
			assert.Equal(u.Hostname()+":8080", address)
			assert.Equal(to, timeout)
			return nil, nil
		},
		DatadogClient: dd,
		DatadogTags: []string{
			"dependency:test-service",
		},
	}
	c, err := checkers.NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status(context.TODO())
	assert.NoError(err)
	assert.Equal(0, dd.IncrCallCount())
}

func TestReachableError(t *testing.T) {
	assert := assert.New(t)
	u, _ := url.Parse("http://example.com")
	cfg := &checkers.ReachableConfig{
		URL: u,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			return nil, errors.New("Failed check")
		},
	}
	c, err := checkers.NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status(context.TODO())
	assert.Error(err)
}

func TestReachableConnError(t *testing.T) {
	assert := assert.New(t)
	u, _ := url.Parse("http://example.com")
	expectedErr := errors.New("Failed close")
	cfg := &checkers.ReachableConfig{
		URL: u,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			conn := &netfakes.FakeConn{}
			conn.CloseReturns(expectedErr)
			return conn, nil
		},
	}
	c, err := checkers.NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status(context.TODO())
	assert.EqualError(err, expectedErr.Error())
}

func TestReachableErrorWithDatadog(t *testing.T) {
	assert := assert.New(t)
	dd := &fakes.FakeReachableDatadogIncrementer{}
	ddTags := []string{
		"dependency:test-service",
	}
	u, _ := url.Parse("http://example.com")
	cfg := &checkers.ReachableConfig{
		URL: u,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			return nil, errors.New("Failed check")
		},
		DatadogClient: dd,
		DatadogTags:   ddTags,
	}
	c, err := checkers.NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status(context.TODO())
	assert.Error(err)
	assert.Equal(1, dd.IncrCallCount())
	name, tags, num := dd.IncrArgsForCall(0)
	assert.Equal(checkers.ReachableDDHealthErrors, name)
	assert.Equal(ddTags, tags)
	assert.Equal(1.0, num)
}
