package checkers

import (
	"fmt"
	"github.com/globalsign/mgo"
)

type MongoConfig struct {
	Auth       *MongoAuthConfig
	Collection string
	DB         string
	Ping       bool
}

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

	session, err := mgo.Dial(cfg.Auth.Url)
	if err != nil {
		return nil, err
	}

	if err := session.Ping(); err != nil {
		return nil, fmt.Errorf("unable to establish initial connection to mongodb: %v", err)
	}

	return &Mongo{
		Config: cfg,
		Session: session,
	}, nil
}

func (m *Mongo) Status() (interface{}, error) {
	if m.Config.Ping {
		fmt.Printf("Checking ping")
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

	if !cfg.Ping && cfg.Collection == "" {
		return fmt.Errorf("At minimum, either cfg.Ping or cfg.Collection")
	}

	if _, err := mgo.ParseURL(cfg.Auth.Url); err != nil {
		return fmt.Errorf("Unable to parse URL: %v", err)
	}

	return nil
}

