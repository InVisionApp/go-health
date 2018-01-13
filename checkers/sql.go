package checkers

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
)

//go:generate counterfeiter -o ../fakes/ipingable.go . IPingable

// IPingable is an interface that allows direct pinging
// of the database by primitive sql.conn objects
type IPingable interface {
	Ping() error
}

// SQLConfig is used for configuring a database check.  The DB field is required.
//
// DB MUST implement either the Pinger interface in the sql/driver package,
// or the IPingable interface in this package.
//
// For more detail, see the following documents:
// 	https://golang.org/pkg/database/sql/#DB.Ping
// 	https://golang.org/pkg/database/sql/driver/#Pinger
//
// Almost all sql libraries/ORM implementations implement one of these two
// interfaces, usually by embedding a native sql Conn.  Driver libraries
// (e.g. MySQL, Postgres, etc) will likely implement the Pinger interface.
type SQLConfig struct {
	DB               interface{} // Required
	implementsPinger bool        // Flag to determine if DB implements driver.Pinger. Set internally
}

// SQL implements the "ICheckable" interface
type SQL struct {
	Config *SQLConfig
}

// NewSQL creates a new database checker that can be used for ".AddCheck(s)".
func NewSQL(cfg *SQLConfig) (*SQL, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Passed in config cannot be nil")
	}
	if cfg.DB == nil {
		return nil, fmt.Errorf("DB interface cannot be nil")
	}

	// this error is returned upon interface validation failure
	badImplementationErr := fmt.Errorf("DB must implement either the " +
		"Pinger interface in the stdlib sql/driver package or the IPingable " +
		"interface in the github.com/InVisionApp/go-health/checkers package")

	dbType := reflect.TypeOf(cfg.DB)
	method, ok := dbType.MethodByName("Ping")
	if !ok {
		return nil, badImplementationErr
	}

	// get number of args passed into ping function. Number is always "self" + 1
	numArgs := method.Type.NumIn()
	switch numArgs {
	case 1:
		// implements IPingable
		// pass
	case 2:
		// implements sql/driver Pinger
		cfg.implementsPinger = true
		c := method.Type.In(1)
		if !(c.PkgPath() == "context" && c.Name() == "Context") {
			return nil, badImplementationErr
		}
	default:
		return nil, badImplementationErr
	}

	// make sure the method returns an error
	if method.Type.NumOut() != 1 {
		return nil, badImplementationErr
	}
	c := method.Type.Out(0)
	if !(c.PkgPath() == "" && c.Name() == "error") {
		return nil, badImplementationErr
	}

	return &SQL{
		Config: cfg,
	}, nil
}

// Status is used for performing a database ping against a dependency; it satisfies
// the "ICheckable" interface.
func (s *SQL) Status() (interface{}, error) {
	if s.Config.implementsPinger {
		db := s.Config.DB.(driver.Pinger)
		ctx := context.Background()
		return nil, db.Ping(ctx)
	}

	// if DB does not implement IPingable, this WILL panic
	db := s.Config.DB.(IPingable)
	return nil, db.Ping()
}
