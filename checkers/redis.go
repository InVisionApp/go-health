package checkers

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const (
	// RedisDefaultSetValue will be used if the "Set" check method is enabled
	// and "RedisSetOptions.Value" is _not_ set.
	RedisDefaultSetValue = "go-health/redis-check"
)

// RedisConfig is used for configuring the go-redis check.
//
// "Auth" is _required_; redis connection/auth config.
//
// "Ping" is optional; the most basic check method, performs a `.Ping()` on the client.
//
// "Get" is optional; perform a "GET" on a key; refer to the "RedisGetOptions" docs for details.
//
// "Set" is optional; perform a "SET" on a key; refer to the "RedisSetOptions" docs for details.
//
// Note: At least _one_ check method must be set/enabled; you can also enable
// _all_ of the check methods (ie. perform a ping, set this key and now try to
// retrieve that key).
type RedisConfig struct {
	Auth *RedisAuthConfig
	Ping bool
	Set  *RedisSetOptions
	Get  *RedisGetOptions
}

// RedisAuthConfig defines how to connect to redis.
type RedisAuthConfig struct {
	Addr     string // `host:port` format
	Password string // leave blank if no password
	DB       int    // leave unset if no specific db

	TLS *tls.Config // TLS config in case we are using in-transit encryption
}

// RedisSetOptions contains attributes that can alter the behavior of the redis
// "SET" check.
//
// "Key" is _required_; the name of the key we are attempting to "SET".
//
// "Value" is optional; what the value should hold; if not set, it will be set
// to "RedisDefaultSetValue".
//
// "Expiration" is optional; if set, a TTL will be attached to the key.
type RedisSetOptions struct {
	Key        string
	Value      string
	Expiration time.Duration
}

// RedisGetOptions contains attributes that can alter the behavior of the redis
// "GET" check.
//
// "Key" is _required_; the name of the key that we are attempting to "GET".
//
// "Expect" is optional; optionally verify that the value for the key matches
// the Expect value.
//
// "NoErrorMissingKey" is optional; by default, the "GET" check will error if
// the key we are fetching does not exist; flip this bool if that is normal/expected/ok.
type RedisGetOptions struct {
	Key               string
	Expect            string
	NoErrorMissingKey bool
}

// Redis implements the ICheckable interface
type Redis struct {
	Config *RedisConfig
	client *redis.Client
}

// NewRedis creates a new "go-redis/redis" checker that can be used w/ "AddChecks()".
func NewRedis(cfg *RedisConfig) (*Redis, error) {
	// validate settings
	if err := validateRedisConfig(cfg); err != nil {
		return nil, fmt.Errorf("Unable to validate redis config: %v", err)
	}

	// try to connect
	c := redis.NewClient(&redis.Options{
		Addr:     cfg.Auth.Addr,
		Password: cfg.Auth.Password, // no password set
		DB:       cfg.Auth.DB,       // use default DB

		TLSConfig: cfg.Auth.TLS,
	})

	if _, err := c.Ping().Result(); err != nil {
		return nil, fmt.Errorf("Unable to establish initial connection to redis: %v", err)
	}

	return &Redis{
		Config: cfg,
		client: c,
	}, nil
}

// Status is used for performing a redis check against a dependency; it satisfies
// the "ICheckable" interface.
func (r *Redis) Status() (interface{}, error) {
	if r.Config.Ping {
		if _, err := r.client.Ping().Result(); err != nil {
			return nil, fmt.Errorf("Ping failed: %v", err)
		}
	}

	if r.Config.Set != nil {
		err := r.client.Set(r.Config.Set.Key, r.Config.Set.Value, r.Config.Set.Expiration).Err()
		if err != nil {
			return nil, fmt.Errorf("Unable to complete set: %v", err)
		}
	}

	if r.Config.Get != nil {
		val, err := r.client.Get(r.Config.Get.Key).Result()
		if err != nil {
			if err == redis.Nil {
				if !r.Config.Get.NoErrorMissingKey {
					return nil, fmt.Errorf("Unable to complete get: '%v' not found", r.Config.Get.Key)
				}
			} else {
				return nil, fmt.Errorf("Unable to complete get: %v", err)
			}
		}

		if r.Config.Get.Expect != "" {
			if r.Config.Get.Expect != val {
				return nil, fmt.Errorf("Unable to complete get: returned value '%v' does not match expected value '%v'",
					val, r.Config.Get.Expect)
			}
		}
	}

	return nil, nil
}

func validateRedisConfig(cfg *RedisConfig) error {
	if cfg == nil {
		return fmt.Errorf("Main config cannot be nil")
	}

	if cfg.Auth == nil {
		return fmt.Errorf("Auth config cannot be nil")
	}

	if cfg.Auth.Addr == "" {
		return fmt.Errorf("Addr string must be set in auth config")
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
			cfg.Set.Value = RedisDefaultSetValue
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
