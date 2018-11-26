package memcachechk

import (
	"bytes"
	"fmt"
	"net"
	"net/url"

	"github.com/bradfitz/gomemcache/memcache"
)

const (
	// MemcachedDefaultSetValue will be used if the "Set" check method is enabled
	// and "MemcachedSetOptions.Value" is _not_ set.
	MemcachedDefaultSetValue = "go-health/memcached-check"
)

// MongoConfig is used for configuring the go-mongo check.
//
// "Url" is _required_; memcached connection url, format is "10.0.0.1:11011". Port (:11011) is mandatory
// "Timeout" defines timeout for socket write/read (useful for servers hosted on different machine)
// "Ping" is optional; Ping establishes tcp connection to memcached server.
type MemcachedConfig struct {
	Url     string
	Timeout int32
	Ping    bool
	Set     *MemcachedSetOptions
	Get     *MemcachedGetOptions
}

type MemcachedClient interface {
	Get(key string) (item *memcache.Item, err error)
	Set(item *memcache.Item) error
}

type Memcached struct {
	Config  *MemcachedConfig
	wrapper *MemcachedClientWrapper
}

// MemcachedSetOptions contains attributes that can alter the behavior of the memcached
// "SET" check.
//
// "Key" is _required_; the name of the key we are attempting to "SET".
//
// "Value" is optional; what the value should hold; if not set, it will be set
// to "MemcachedDefaultSetValue".
//
// "Expiration" is optional; if set, a TTL will be attached to the key.
type MemcachedSetOptions struct {
	Key        string
	Value      string
	Expiration int32
}

// MemcachedGetOptions contains attributes that can alter the behavior of the memcached
// "GET" check.
//
// "Key" is _required_; the name of the key that we are attempting to "GET".
//
// "Expect" is optional; optionally verify that the value for the key matches
// the Expect value.
//
// "NoErrorMissingKey" is optional; by default, the "GET" check will error if
// the key we are fetching does not exist; flip this bool if that is normal/expected/ok.
type MemcachedGetOptions struct {
	Key               string
	Expect            []byte
	NoErrorMissingKey bool
}

func NewMemcached(cfg *MemcachedConfig) (*Memcached, error) {
	// validate settings
	if err := validateMemcachedConfig(cfg); err != nil {
		return nil, fmt.Errorf("unable to validate memcached config: %v", err)
	}

	mcWrapper := &MemcachedClientWrapper{memcache.New(cfg.Url)}

	return &Memcached{
		Config:  cfg,
		wrapper: mcWrapper,
	}, nil
}

func (mc *Memcached) Status() (interface{}, error) {

	if mc.Config.Ping {
		if _, err := net.Dial("tcp", mc.Config.Url); err != nil {
			return nil, fmt.Errorf("Ping failed: %v", err)
		}
	}

	if mc.Config.Set != nil {
		err := mc.wrapper.GetClient().Set(&memcache.Item{Key: mc.Config.Set.Key, Value: []byte(mc.Config.Set.Value), Expiration: mc.Config.Set.Expiration})
		if err != nil {
			return nil, fmt.Errorf("Unable to complete set: %v", err)
		}
	}

	if mc.Config.Get != nil {
		val, err := mc.wrapper.GetClient().Get(mc.Config.Get.Key)
		if err != nil {
			if err == memcache.ErrCacheMiss {
				if !mc.Config.Get.NoErrorMissingKey {
					return nil, fmt.Errorf("Unable to complete get: '%v' not found", mc.Config.Get.Key)
				}
			} else {
				return nil, fmt.Errorf("Unable to complete get: %v", err)
			}
		}

		if mc.Config.Get.Expect != nil {
			if !bytes.Equal(mc.Config.Get.Expect, val.Value) {
				return nil, fmt.Errorf("Unable to complete get: returned value '%v' does not match expected value '%v'",
					val, mc.Config.Get.Expect)
			}
		}
	}

	return nil, nil
}

func validateMemcachedConfig(cfg *MemcachedConfig) error {
	if cfg == nil {
		return fmt.Errorf("Main config cannot be nil")
	}

	if cfg.Url == "" {
		return fmt.Errorf("Url string must be set in config")
	}

	if _, err := url.Parse(cfg.Url); err != nil {
		return fmt.Errorf("Unable to parse URL: %v", err)
	}

	// At least one check method must be set
	if !cfg.Ping && cfg.Set == nil && cfg.Get == nil {
		return fmt.Errorf("At minimum, either cfg.Ping, cfg.Set or cfg.Get must be set")
	}

	// If .Set is set, verify that at minimum .Key is set
	if cfg.Set != nil {
		if cfg.Set.Key == "" {
			return fmt.Errorf("If cfg.Set is used, cfg.Set.Key must be set")
		}

		if cfg.Set.Value == "" {
			cfg.Set.Value = MemcachedDefaultSetValue
		}
	}

	// If .Get is set, verify that at minimum .Key is set
	if cfg.Get != nil {
		if cfg.Get.Key == "" {
			return fmt.Errorf("If cfg.Get is used, cfg.Get.Key must be set")
		}
	}

	return nil
}

// Used to simplify testing routines
type MemcachedClientWrapper struct {
	MemcachedClient
}

func (mcw MemcachedClientWrapper) GetClient() MemcachedClient {
	return mcw.MemcachedClient
}
