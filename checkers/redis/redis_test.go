package redischk

import (
	"fmt"
	"testing"

	"github.com/alicebob/miniredis"
	. "github.com/onsi/gomega"
)

func TestNewRedis(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		server, err := miniredis.Run()
		Expect(err).ToNot(HaveOccurred())

		cfg := &RedisConfig{
			Ping: true,
			Auth: &RedisAuthConfig{
				Addr: server.Addr(),
			},
		}

		r, err := NewRedis(cfg)

		Expect(err).ToNot(HaveOccurred())
		Expect(r).ToNot(BeNil())
	})

	t.Run("Bad config should error", func(t *testing.T) {
		var cfg *RedisConfig
		r, err := NewRedis(cfg)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unable to validate redis config"))
		Expect(r).To(BeNil())
	})

	t.Run("Should error when redis server is not available", func(t *testing.T) {
		cfg := &RedisConfig{
			Ping: true,
			Auth: &RedisAuthConfig{
				Addr: "foobar:42848",
			},
		}

		r, err := NewRedis(cfg)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unable to establish"))
		Expect(r).To(BeNil())
	})
}

func TestValidateRedisConfig(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error with nil main config", func(t *testing.T) {
		var cfg *RedisConfig
		err := validateRedisConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Main config cannot be nil"))
	})

	t.Run("Should error with nil auth config", func(t *testing.T) {
		err := validateRedisConfig(&RedisConfig{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Auth config cannot be nil"))
	})

	t.Run("Auth config must have an addr set", func(t *testing.T) {
		cfg := &RedisConfig{
			Auth: &RedisAuthConfig{},
		}

		err := validateRedisConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Addr string must be set"))
	})

	t.Run("Should error if none of the check methods are enabled", func(t *testing.T) {
		cfg := &RedisConfig{
			Auth: &RedisAuthConfig{
				Addr: "localhost:6379",
			},
		}

		err := validateRedisConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("At minimum, either cfg.Ping, cfg.Set or cfg.Get"))
	})

	t.Run("Should error if .Set is used but key is undefined", func(t *testing.T) {
		cfg := &RedisConfig{
			Auth: &RedisAuthConfig{
				Addr: "localhost:6379",
			},
			Set: &RedisSetOptions{},
		}

		err := validateRedisConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("If cfg.Set is used, cfg.Set.Key must be set"))
	})

	t.Run("Should error if .Get is used but key is undefined", func(t *testing.T) {
		cfg := &RedisConfig{
			Auth: &RedisAuthConfig{
				Addr: "localhost:6379",
			},
			Get: &RedisGetOptions{},
		}

		err := validateRedisConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("If cfg.Get is used, cfg.Get.Key must be set"))
	})

	t.Run("If Set is enabled but value is unset, should use default value", func(t *testing.T) {
		cfg := &RedisConfig{
			Auth: &RedisAuthConfig{
				Addr: "localhost:6379",
			},
			Set: &RedisSetOptions{
				Key: "foo",
			},
		}

		err := validateRedisConfig(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.Set.Value).To(Equal(RedisDefaultSetValue))
	})
}

func TestRedisStatus(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error when ping is enabled and fails", func(t *testing.T) {
		cfg := &RedisConfig{
			Ping: true,
		}
		checker, server, err := setupRedis(cfg)
		if err != nil {
			t.Fatal(err)
		}

		// Stop the server, so ping check fails
		server.Close()

		Expect(err).ToNot(HaveOccurred())

		_, err = checker.Status()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Ping failed"))
	})

	t.Run("When set is enabled", func(t *testing.T) {
		t.Run("should error if set fails", func(t *testing.T) {
			cfg := &RedisConfig{
				Set: &RedisSetOptions{
					Key: "test-key",
				},
			}
			checker, server, err := setupRedis(cfg)
			if err != nil {
				t.Fatal(err)
			}

			// Stop the server, so ping check fails
			server.Close()

			_, err = checker.Status()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unable to complete set"))
		})

		t.Run("should use .Value if .Value is defined", func(t *testing.T) {
			cfg := &RedisConfig{
				Set: &RedisSetOptions{
					Key:   "test-key",
					Value: "test-value",
				},
			}
			checker, server, err := setupRedis(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status()
			Expect(err).ToNot(HaveOccurred())

			val, err := server.Get(cfg.Set.Key)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(cfg.Set.Value))
		})

		t.Run("should use default .Value if .Value is not explicitly set", func(t *testing.T) {
			cfg := &RedisConfig{
				Set: &RedisSetOptions{
					Key: "test-key",
				},
			}
			checker, server, err := setupRedis(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status()
			Expect(err).ToNot(HaveOccurred())

			val, err := server.Get(cfg.Set.Key)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(RedisDefaultSetValue))
		})
	})

	t.Run("When get is enabled", func(t *testing.T) {
		t.Run("should error if key is missing and NoErrorMissingKey not set", func(t *testing.T) {
			cfg := &RedisConfig{
				Get: &RedisGetOptions{
					Key: "test-key",
				},
			}
			checker, server, err := setupRedis(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Unable to complete get: '%v' not found", cfg.Get.Key)))
		})

		t.Run("should NOT error if key is missing and NoErrorMissingKey IS set", func(t *testing.T) {
			cfg := &RedisConfig{
				Get: &RedisGetOptions{
					Key:               "test-key",
					NoErrorMissingKey: true,
				},
			}
			checker, server, err := setupRedis(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status()
			Expect(err).ToNot(HaveOccurred())
		})

		t.Run("should error if get fails", func(t *testing.T) {
			cfg := &RedisConfig{
				Get: &RedisGetOptions{
					Key:               "test-key",
					NoErrorMissingKey: true,
				},
			}
			checker, server, err := setupRedis(cfg)
			if err != nil {
				t.Fatal(err)
			}

			// Close the server so the GET fails
			server.Close()

			_, err = checker.Status()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unable to complete get"))
		})

		t.Run("should error if .Expect is set and the value does not match", func(t *testing.T) {
			cfg := &RedisConfig{
				Set: &RedisSetOptions{
					Key:   "test-key",
					Value: "foo",
				},
				Get: &RedisGetOptions{
					Key:               "test-key",
					Expect:            "bar",
					NoErrorMissingKey: true,
				},
			}
			checker, server, err := setupRedis(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not match expected value"))
		})

		t.Run("should NOT error if .Expect is not set", func(t *testing.T) {
			cfg := &RedisConfig{
				Set: &RedisSetOptions{
					Key:   "test-key",
					Value: "foo",
				},
				Get: &RedisGetOptions{
					Key: "test-key",
				},
			}
			checker, server, err := setupRedis(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status()
			Expect(err).ToNot(HaveOccurred())
		})
	})
}

func setupRedis(cfg *RedisConfig) (*Redis, *miniredis.Miniredis, error) {
	server, err := miniredis.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to setup miniredis: %v", err)
	}

	cfg.Auth = &RedisAuthConfig{
		Addr: server.Addr(),
	}

	checker, err := NewRedis(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to setup checker: %v", err)
	}

	return checker, server, nil
}
