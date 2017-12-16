package checkers

import (
	"testing"

	"github.com/go-redis/redis"
	. "github.com/onsi/gomega"
)

func TestNewRedis(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})

		cfg := &RedisConfig{
			Client: client,
			Ping:   true,
		}

		r, err := NewRedis(cfg)

		Expect(err).ToNot(HaveOccurred())
		Expect(r).ToNot(BeNil())
	})

	t.Run("Bad config should error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})

		cfg := &RedisConfig{
			Client: client,
		}

		r, err := NewRedis(cfg)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unable to validate redis config"))
		Expect(r).To(BeNil())

	})
}

func TestValidateRedisConfig(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error with no client", func(t *testing.T) {
		err := validateRedisConfig(&RedisConfig{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Client cannot be nil"))
	})

	t.Run("Should error if none of the check methods are enabled", func(t *testing.T) {
		cfg := &RedisConfig{
			Client: redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			}),
		}

		err := validateRedisConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("At minimum, either cfg.Ping, cfg.Set or cfg.Get"))
	})

	t.Run("Should error if .Set is used but key is undefined", func(t *testing.T) {
		cfg := &RedisConfig{
			Client: redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			}),
			Set: &RedisSetOptions{},
		}

		err := validateRedisConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("If cfg.Set is used, cfg.Set.Key must be set"))
	})

	t.Run("Should error if .Get is used but key is undefined", func(t *testing.T) {
		cfg := &RedisConfig{
			Client: redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			}),
			Get: &RedisGetOptions{},
		}

		err := validateRedisConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("If cfg.Get is used, cfg.Get.Key must be set"))
	})

	t.Run("If Set is enabled but value is unset, should use default value", func(t *testing.T) {
		cfg := &RedisConfig{
			Client: redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			}),
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

	})

	t.Run("When set is enabled", func(t *testing.T) {
		t.Run("should error if set fails", func(t *testing.T) {

		})

		t.Run("should use .Value if .Value is defined", func(t *testing.T) {

		})
	})

	t.Run("When get is enabled", func(t *testing.T) {
		t.Run("should error if key is missing and NoErrorMissingKey not set", func(t *testing.T) {

		})

		t.Run("should NOT error if key is missing and NoErrorMissingKey IS set", func(t *testing.T) {

		})

		t.Run("should error if get fails", func(t *testing.T) {

		})

		t.Run("should error if .Expect is set and the value does not match", func(t *testing.T) {

		})

		t.Run("should NOT error if .Expect is not set", func(t *testing.T) {

		})
	})
}
