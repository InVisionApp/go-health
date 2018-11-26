package mongochk

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
)

const (
	DefaultDialTimeout = 10 * time.Second
)

// MongoConfig is used for configuring the go-mongo check.
//
// "Auth" is _required_; redis connection/auth config.
//
// "Collection" is optional; method checks if collection exist
//
// "Ping" is optional; Ping runs a trivial ping command just to get in touch with the server.
//
// "DialTimeout" is optional; default @ 10s; determines the max time we'll wait to reach a server.
//
// Note: At least _one_ check method must be set/enabled; you can also enable
// _all_ of the check methods (ie. perform a ping, or check particular collection for existense).
type MongoConfig struct {
	Auth        *MongoAuthConfig
	Collection  string
	DB          string
	Ping        bool
	DialTimeout time.Duration
}

// MongoAuthConfig, used to setup connection params for go-mongo check
// Url format is localhost:27017 or mongo://localhost:27017
// Credential has format described at https://godoc.org/github.com/globalsign/mgo#Credential
type MongoAuthConfig struct {
	Url         string
	Credentials mgo.Credential
}

type Mongo struct {
	Config  *MongoConfig
	Session *mgo.Session
}

func NewMongo(cfg *MongoConfig) (*Mongo, error) {
	// validate settings
	if err := validateMongoConfig(cfg); err != nil {
		return nil, fmt.Errorf("unable to validate mongodb config: %v", err)
	}

	session, err := mgo.DialWithTimeout(cfg.Auth.Url, cfg.DialTimeout)
	if err != nil {
		return nil, err
	}

	if err := session.Ping(); err != nil {
		return nil, fmt.Errorf("unable to establish initial connection to mongodb: %v", err)
	}

	return &Mongo{
		Config:  cfg,
		Session: session,
	}, nil
}

func (m *Mongo) Status() (interface{}, error) {
	if m.Config.Ping {
		if err := m.Session.Ping(); err != nil {
			return nil, fmt.Errorf("ping failed: %v", err)
		}
	}

	if m.Config.Collection != "" {
		collections, err := m.Session.DB(m.Config.DB).CollectionNames()
		if err != nil {
			return nil, fmt.Errorf("unable to complete set: %v", err)
		}
		if !contains(collections, m.Config.Collection) {
			return nil, fmt.Errorf("mongo db %v collection not found", m.Config.Collection)
		}
	}

	return nil, nil
}

func contains(data []string, needle string) bool {
	for _, item := range data {
		if item == needle {
			return true
		}
	}
	return false
}

func validateMongoConfig(cfg *MongoConfig) error {
	if cfg == nil {
		return fmt.Errorf("Main config cannot be nil")
	}

	if cfg.Auth == nil {
		return fmt.Errorf("Auth config cannot be nil")
	}

	if cfg.Auth.Url == "" {
		return fmt.Errorf("Url string must be set in auth config")
	}

	if _, err := mgo.ParseURL(cfg.Auth.Url); err != nil {
		return fmt.Errorf("Unable to parse URL: %v", err)
	}

	if !cfg.Ping && cfg.Collection == "" {
		return fmt.Errorf("At minimum, either cfg.Ping or cfg.Collection")
	}

	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = DefaultDialTimeout
	}

	return nil
}
