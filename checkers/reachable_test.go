package checkers

import (
	"errors"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/InVisionApp/go-health/fakes/netfakes"
	"github.com/InVisionApp/platform-dashboard-web/datadog/datadogfakes"
	"github.com/stretchr/testify/assert"
)

func TestReachableSuccess(t *testing.T) {
	assert := assert.New(t)
	dd := &datadogfakes.FakeIStatsDClient{}
	u, _ := url.Parse("http://example.com")
	cfg := &ReachableConfig{
		URL:           u,
		DatadogClient: dd,
		DatadogTags: []string{
			"dependency:test-service",
		},
	}
	c, err := NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)
	c.dialer = func(network, address string, timeout time.Duration) (net.Conn, error) {
		return nil, nil
	}

	_, err = c.Status()
	assert.NoError(err)
	assert.Equal(0, dd.IncrCallCount())
	assert.Equal("dependency:test-service", c.tags[0])
}

func TestReachableError(t *testing.T) {
	assert := assert.New(t)
	u, _ := url.Parse("http://example.com")
	cfg := &ReachableConfig{
		URL: u,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			return nil, errors.New("Failed check")
		},
	}
	c, err := NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status()
	assert.Error(err)
}

func TestReachableConnError(t *testing.T) {
	assert := assert.New(t)
	u, _ := url.Parse("http://example.com")
	expectedErr := errors.New("Failed close")
	cfg := &ReachableConfig{
		URL: u,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			conn := &netfakes.FakeConn{}
			conn.CloseReturns(expectedErr)
			return conn, nil
		},
	}
	c, err := NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status()
	assert.EqualError(err, expectedErr.Error())
}

func TestReachableErrorWithDatadog(t *testing.T) {
	assert := assert.New(t)
	dd := &datadogfakes.FakeIStatsDClient{}
	ddTags := []string{
		"dependency:test-service",
	}
	u, _ := url.Parse("http://example.com")
	cfg := &ReachableConfig{
		URL: u,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			return nil, errors.New("Failed check")
		},
		DatadogClient: dd,
		DatadogTags:   ddTags,
	}
	c, err := NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status()
	assert.Error(err)
	assert.Equal(1, dd.IncrCallCount())
	name, tags, num := dd.IncrArgsForCall(0)
	assert.Equal(ReachableDDHealthErrors, name)
	assert.Equal(ddTags, tags)
	assert.Equal(1.0, num)
}
