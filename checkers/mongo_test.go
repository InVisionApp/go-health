package checkers

import (
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/zaffka/mongodb-boltdb-mock/db"
	"testing"
)

func TestNewMongo(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		mongo := db.New(&db.Mongo{})
		url := "mongo://localhost:/27017"
		err := mongo.Connect(url)
		defer mongo.Close()

		Expect(err).ToNot(HaveOccurred())

		cfg := &MongoConfig{
			Ping: true,
			Auth: &MongoAuthConfig{
				Url: url,
			},
		}

		r, err := NewMongo(cfg)

		Expect(err).ToNot(HaveOccurred())
		Expect(r).ToNot(BeNil())
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
				Url: "foobar:42848",
			},
		}

		r, err := NewMongo(cfg)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unable to establish"))
		Expect(r).To(BeNil())
	})
}

func TestValidateMongoConfig(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error with nil main config", func(t *testing.T) {
		var cfg *MongoConfig
		err := validateMongoConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Main config cannot be nil"))
	})

	t.Run("Should error with nil auth config", func(t *testing.T) {
		err := validateMongoConfig(&MongoConfig{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Auth config cannot be nil"))
	})

	t.Run("Auth config must have an addr set", func(t *testing.T) {
		cfg := &MongoConfig{
			Auth: &MongoAuthConfig{},
		}

		err := validateMongoConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Addr string must be set"))
	})

	t.Run("Should error if none of the check methods are enabled", func(t *testing.T) {
		cfg := &MongoConfig{
			Auth: &MongoAuthConfig{
				Url: "localhost:6379",
			},
		}

		err := validateMongoConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("At minimum, either cfg.Ping or cfg.Collection"))
	})

}

func TestMongoStatus(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error when ping is enabled and fails", func(t *testing.T) {
		cfg := &MongoConfig{
			Ping: true,
		}
		checker, server, err := setupMongo(cfg)
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

}

func setupMongo(cfg *MongoConfig) (*Mongo, db.Handler, error) {
	server := db.New(&db.Mongo{})
	url := "mongo://localhost:/27017"
	err := server.Connect(url)
	defer server.Close()

	if err != nil {
		return nil, nil, fmt.Errorf("Unable to setup mongo: %v", err)
	}

	cfg.Auth = &MongoAuthConfig{
		Url: url,
	}

	checker, err := NewMongo(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to setup checker: %v", err)
	}

	return checker, server, nil
}
