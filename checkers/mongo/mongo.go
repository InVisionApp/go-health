package mongochk

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
// Url mongodb://localhost:27017
type MongoAuthConfig struct {
	Url string
}

type Mongo struct {
	Config *MongoConfig
	Client *mongo.Client
}

func NewMongo(cfg *MongoConfig) (*Mongo, error) {
	// validate settings
	if err := validateMongoConfig(cfg); err != nil {
		return nil, fmt.Errorf("unable to validate mongodb config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Auth.Url))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("unable to establish initial connection to mongodb: %v", err)
	}

	return &Mongo{
		Config: cfg,
		Client: client,
	}, nil
}

func (m *Mongo) Status(ctx context.Context) (interface{}, error) {
	if m.Config.Ping {
		if err := m.Client.Ping(ctx, nil); err != nil {
			return nil, fmt.Errorf("ping failed: %v", err)
		}
	}

	if m.Config.DB != "" && m.Config.Collection != "" {
		cur, err := m.Client.Database(m.Config.DB).
			ListCollections(ctx, bson.D{{"name", m.Config.Collection}}, options.ListCollections().SetNameOnly(true))

		if err != nil {
			return nil, fmt.Errorf("unable to complete set: %v", err)
		}

		defer cur.Close(ctx)

		if !cur.Next(ctx) {
			if err := cur.Err(); err != nil {
				return nil, err
			}

			return nil, fmt.Errorf("mongo db %v collection not found", m.Config.Collection)
		}
	}

	return nil, nil
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

	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = DefaultDialTimeout
	}

	return nil
}
