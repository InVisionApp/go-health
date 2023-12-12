package memcachechk

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	. "github.com/onsi/gomega"
)

const (
	testUrl = "localhost:11011"
)

var emulateServerShutdown bool

func TestNewMemcached(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		url := testUrl
		cfg := &MemcachedConfig{
			Url:  url,
			Ping: true,
		}
		mc, server, err := setupMemcached(cfg)

		Expect(err).ToNot(HaveOccurred())
		Expect(mc).ToNot(BeNil())
		server.Close()
	})

	t.Run("Bad config should error", func(t *testing.T) {
		var cfg *MemcachedConfig
		mc, err := NewMemcached(cfg)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unable to validate memcached config"))
		Expect(mc).To(BeNil())
	})

	t.Run("Memcached should contain Client and Config", func(t *testing.T) {
		url := testUrl
		cfg := &MemcachedConfig{
			Url:  url,
			Ping: true,
		}
		mc, err := NewMemcached(cfg)

		Expect(err).ToNot(HaveOccurred())
		Expect(mc).ToNot(BeNil())
	})

}

func TestValidateMemcachedConfig(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error with nil main config", func(t *testing.T) {
		var cfg *MemcachedConfig
		err := validateMemcachedConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Main config cannot be nil"))
	})

	t.Run("Config must have an url set", func(t *testing.T) {
		cfg := &MemcachedConfig{}

		err := validateMemcachedConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Url string must be set in config"))
	})

	t.Run("Should error if none of the check methods are enabled", func(t *testing.T) {
		cfg := &MemcachedConfig{
			Url: testUrl,
		}

		err := validateMemcachedConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("At minimum, either cfg.Ping, cfg.Set or cfg.Get must be set"))
	})

	t.Run("Should error if .Set is used but key is undefined", func(t *testing.T) {
		cfg := &MemcachedConfig{
			Url: testUrl,
			Set: &MemcachedSetOptions{},
		}

		err := validateMemcachedConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("If cfg.Set is used, cfg.Set.Key must be set"))
	})

	t.Run("Should error if .Get is used but key is undefined", func(t *testing.T) {
		cfg := &MemcachedConfig{
			Url: testUrl,
			Get: &MemcachedGetOptions{},
		}

		err := validateMemcachedConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("If cfg.Get is used, cfg.Get.Key must be set"))
	})

	t.Run("Should error if url has wrong format", func(t *testing.T) {
		cfg := &MemcachedConfig{
			Url: "wrong\\localhost:6379",
		}

		err := validateMemcachedConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unable to parse URL"))
	})

	t.Run("Shouldn't error with properly set config", func(t *testing.T) {
		cfg := &MemcachedConfig{
			Url: testUrl,
			Get: &MemcachedGetOptions{
				Key:    "should_return_valid",
				Expect: []byte("should_return_valid"),
			},
		}
		err := validateMemcachedConfig(cfg)
		Expect(err).To(BeNil())
	})

}

func TestMemcachedStatus(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error when ping is enabled", func(t *testing.T) {
		cfg := &MemcachedConfig{
			Ping: true,
		}
		checker, _, err := setupMemcached(cfg)
		if err != nil {
			t.Fatal(err)
		}
		_, err = checker.Status(context.TODO())
		Expect(err).To(HaveOccurred())

		_, err = checker.Status(context.TODO())
		Expect(err.Error()).To(ContainSubstring("Ping failed"))
	})

	t.Run("When set is enabled", func(t *testing.T) {
		t.Run("should error if set fails", func(t *testing.T) {
			cfg := &MemcachedConfig{
				Set: &MemcachedSetOptions{
					Key: "valid",
				},
			}
			checker, server, err := setupMemcached(cfg)
			if err != nil {
				t.Fatal(err)
			}

			// Mark server is stoppped
			server.Close()

			_, err = checker.Status(context.TODO())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unable to complete set"))
		})

		t.Run("should use .Value if .Value is defined", func(t *testing.T) {
			cfg := &MemcachedConfig{
				Set: &MemcachedSetOptions{
					Key:   "valid",
					Value: "valid",
				},
			}
			checker, server, err := setupMemcached(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status(context.TODO())
			Expect(err).ToNot(HaveOccurred())

			val, err := checker.wrapper.GetClient().Get(cfg.Set.Key)
			Expect(err).ToNot(HaveOccurred())
			Expect(val.Value).To(Equal([]byte(cfg.Set.Value)))
		})

		t.Run("should use default .Value if .Value is not explicitly set", func(t *testing.T) {
			cfg := &MemcachedConfig{
				Set: &MemcachedSetOptions{
					Key: "should_return_default",
				},
			}
			checker, server, err := setupMemcached(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status(context.TODO())
			Expect(err).ToNot(HaveOccurred())

			val, err := checker.wrapper.GetClient().Get(cfg.Set.Key)
			Expect(err).ToNot(HaveOccurred())
			Expect(val.Value).To(Equal([]byte(MemcachedDefaultSetValue)))
		})

		t.Run("should use default .Value if .Value is set to empty string", func(t *testing.T) {
			cfg := &MemcachedConfig{
				Set: &MemcachedSetOptions{
					Key:   "should_return_default",
					Value: "",
				},
			}
			checker, server, err := setupMemcached(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status(context.TODO())
			Expect(err).ToNot(HaveOccurred())

			val, err := checker.wrapper.GetClient().Get(cfg.Set.Key)
			Expect(err).ToNot(HaveOccurred())
			Expect(val.Value).To(Equal([]byte(MemcachedDefaultSetValue)))
		})
	})

	t.Run("When get is enabled", func(t *testing.T) {
		t.Run("should error if key is missing and NoErrorMissingKey not set", func(t *testing.T) {
			cfg := &MemcachedConfig{
				Get: &MemcachedGetOptions{
					Key: "should_return_error",
				},
			}
			checker, server, err := setupMemcached(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status(context.TODO())
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Unable to complete get: '%v' not found", cfg.Get.Key)))
		})

		t.Run("should NOT error if key is missing and NoErrorMissingKey IS set", func(t *testing.T) {
			cfg := &MemcachedConfig{
				Get: &MemcachedGetOptions{
					Key:               "should_return_valid",
					NoErrorMissingKey: true,
				},
			}
			checker, server, err := setupMemcached(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status(context.TODO())
			Expect(err).ToNot(HaveOccurred())
		})

		t.Run("should error if get fails", func(t *testing.T) {
			cfg := &MemcachedConfig{
				Get: &MemcachedGetOptions{
					Key:               "anything_here",
					NoErrorMissingKey: true,
				},
			}
			checker, server, err := setupMemcached(cfg)
			if err != nil {
				t.Fatal(err)
			}

			// Close the server so the GET fails
			server.Close()

			_, err = checker.Status(context.TODO())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unable to complete get"))
		})

		t.Run("should error if .Expect is set and the value does not match", func(t *testing.T) {
			cfg := &MemcachedConfig{
				Set: &MemcachedSetOptions{
					Key:   "should_return_invalid",
					Value: "foo",
				},
				Get: &MemcachedGetOptions{
					Key:               "should_return_invalid",
					Expect:            []byte("bar"),
					NoErrorMissingKey: true,
				},
			}
			checker, server, err := setupMemcached(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status(context.TODO())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not match expected value"))
		})

		t.Run("should NOT error if .Expect is not set", func(t *testing.T) {
			cfg := &MemcachedConfig{
				Set: &MemcachedSetOptions{
					Key:   "test-key",
					Value: "foo",
				},
				Get: &MemcachedGetOptions{
					Key: "test-key",
				},
			}
			checker, server, err := setupMemcached(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			_, err = checker.Status(context.TODO())
			Expect(err).ToNot(HaveOccurred())
		})
	})

}

func setupMemcached(cfg *MemcachedConfig) (*Memcached, *MockServer, error) {
	server := &MockServer{}
	server.Reset()
	cfg.Url = testUrl
	checker := &Memcached{
		wrapper: &MemcachedClientWrapper{&MockMemcachedClient{}},
		Config:  cfg,
	}

	return checker, server, nil
}

type MockServer struct{}

func (s *MockServer) Close() {
	emulateServerShutdown = true
}

func (s *MockServer) Reset() {
	emulateServerShutdown = false
}

type MockMemcachedClient struct{}

func (m *MockMemcachedClient) Get(key string) (item *memcache.Item, err error) {
	if emulateServerShutdown {
		return nil, fmt.Errorf("Unable to complete get")
	}
	switch key {
	case "should_return_valid":
		return &memcache.Item{Key: key, Value: []byte(key)}, nil
	case "should_return_invalid":
		return &memcache.Item{Key: key, Value: []byte(key + strconv.Itoa(rand.Int()))}, nil
	case "should_return_default":
		return &memcache.Item{Key: key, Value: []byte(MemcachedDefaultSetValue)}, nil
	case "should_return_error":
		return &memcache.Item{Key: key, Value: []byte(key)}, memcache.ErrCacheMiss
	default:
		return &memcache.Item{Key: key, Value: []byte(key)}, nil
	}
}

func (m *MockMemcachedClient) Set(item *memcache.Item) error {
	if emulateServerShutdown {
		return fmt.Errorf("Unable to complete set")
	}
	return nil
}
