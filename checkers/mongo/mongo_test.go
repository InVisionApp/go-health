package mongochk

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/zaffka/mongodb-boltdb-mock/db"
)

func TestNewMongo(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		mongo := db.New(&db.Mock{})
		url := "localhost:27017"
		err := mongo.Connect(url)
		defer mongo.Close()
		Expect(err).ToNot(HaveOccurred())
	})

	t.Run("Bad config should error", func(t *testing.T) {
		var cfg *MongoConfig
		r, err := NewMongo(cfg)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unable to validate mongodb config"))
		Expect(r).To(BeNil())
	})

	t.Run("Should error when mongo server is not available", func(t *testing.T) {
		cfg := &MongoConfig{
			Ping: true,
			Auth: &MongoAuthConfig{
				URL: "foobar:42848",
			},
			DialTimeout: 20 * time.Millisecond,
		}

		r, err := NewMongo(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no reachable servers"))
		Expect(r).To(BeNil())
	})
}

func TestValidateMongoConfig(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error with nil main config", func(t *testing.T) {
		var cfg *MongoConfig
		err := validateMongoConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("main config cannot be nil"))
	})

	t.Run("Should error with nil auth config", func(t *testing.T) {
		err := validateMongoConfig(&MongoConfig{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("auth config cannot be nil"))
	})

	t.Run("auth config must have an addr set", func(t *testing.T) {
		cfg := &MongoConfig{
			Auth: &MongoAuthConfig{},
		}

		err := validateMongoConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("url string must be set in auth config"))
	})

	t.Run("should error if none of the check methods are enabled", func(t *testing.T) {
		cfg := &MongoConfig{
			Auth: &MongoAuthConfig{
				URL: "localhost:6379",
			},
		}

		err := validateMongoConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("at minimum, either cfg.Ping or cfg.Collection"))
	})

	t.Run("should error if url has wrong format", func(t *testing.T) {
		cfg := &MongoConfig{
			Auth: &MongoAuthConfig{
				URL: "localhost:40001?foo=1&bar=2",
			},
		}

		err := validateMongoConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unable to parse URL"))
	})

}

func TestMongoStatus(t *testing.T) {
	RegisterTestingT(t)

	t.Run("shouldn't error when ping is enabled", func(t *testing.T) {
		cfg := &MongoConfig{
			Ping: true,
		}
		checker, _, err := setupMongo(cfg)
		if err != nil {
			t.Fatal(err)
		}

		Expect(err).ToNot(HaveOccurred())

		_, err = checker.Status()

		Expect(err).To(BeNil())
	})

	t.Run("Should error if collection not found(available)", func(t *testing.T) {
		cfg := &MongoConfig{
			Collection: "go-check",
		}
		checker, _, err := setupMongo(cfg)
		if err != nil {
			t.Fatal(err)
		}

		_, err = checker.Status()

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("collection not found"))
	})

}

func setupMongo(cfg *MongoConfig) (*Mongo, db.Handler, error) {
	server := db.New(&db.Mongo{})
	url := "mongodb://localhost:27017"
	err := server.Connect(url)

	if err != nil {
		return nil, nil, fmt.Errorf("unable to setup mongo: %v", err)
	}

	cfg.Auth = &MongoAuthConfig{
		URL: url,
	}

	checker, err := NewMongo(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to setup checker: %v", err)
	}

	return checker, server, nil
}
