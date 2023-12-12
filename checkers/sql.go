package checkers

import (
	"context"
	"database/sql"
	"fmt"
)

//go:generate counterfeiter -o ../fakes/isqlpinger.go . SQLPinger
//go:generate counterfeiter -o ../fakes/isqlqueryer.go . SQLQueryer
//go:generate counterfeiter -o ../fakes/isqlexecer.go . SQLExecer

// SQLPinger is an interface that allows direct pinging of the database
type SQLPinger interface {
	PingContext(ctx context.Context) error
}

// SQLQueryer is an interface that allows querying of the database
type SQLQueryer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// SQLExecer is an interface that allows executing of queries in the database
type SQLExecer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// SQLQueryerResultHandler is the BYO function to
// handle the result of an SQL SELECT query
type SQLQueryerResultHandler func(rows *sql.Rows) (bool, error)

// SQLExecerResultHandler is the BYO function
// to handle a database exec result
type SQLExecerResultHandler func(result sql.Result) (bool, error)

// SQLConfig is used for configuring a database check.
// One of the Pinger, Queryer, or Execer fields is required.
//
// If Execer is set, it will take precedence over Queryer and Pinger,
// Execer implements the SQLExecer interface in this package.
// The sql.DB and sql.TX structs both implement this interface.
//
// Note that if the Execer is set, then the ExecerResultHandler
// and Query values MUST also be set
//
// If Queryer is set, it will take precedence over Pinger.
// SQLQueryer implements the SQLQueryer interface in this package.
// The sql.DB and sql.TX structs both implement this interface.
//
// Note that if the Queryer is set, then the QueryerResultHandler
// and Query values MUST also be set
//
// Pinger implements the SQLPinger interface in this package.
// The sql.DB struct implements this interface.
type SQLConfig struct {
	// Pinger is the value implementing SQLPinger
	Pinger SQLPinger

	// Queryer is the value implementing SQLQueryer
	Queryer SQLQueryer

	// Execer is the value implementing SQLExecer
	Execer SQLExecer

	// Query is the parameterized SQL query required
	// with both Queryer and Execer
	Query string

	// Params are the SQL query parameters, if any
	Params []interface{}

	// QueryerResultHandler handles the result of
	// the QueryContext function
	QueryerResultHandler SQLQueryerResultHandler

	// ExecerResultHandler handles the result of
	// the ExecContext function
	ExecerResultHandler SQLExecerResultHandler
}

// SQL implements the "ICheckable" interface
type SQL struct {
	Config *SQLConfig
}

// NewSQL creates a new database checker that can be used for ".AddCheck(s)".
func NewSQL(cfg *SQLConfig) (*SQL, error) {
	if err := validateSQLConfig(cfg); err != nil {
		return nil, err
	}

	return &SQL{
		Config: cfg,
	}, nil
}

// DefaultQueryHandler is the default SQLQueryer result handler
// that assumes one row was returned from the passed query
func DefaultQueryHandler(rows *sql.Rows) (bool, error) {
	defer rows.Close()

	numRows := 0
	for rows.Next() {
		numRows++
	}

	return numRows == 1, nil
}

// DefaultExecHandler is the default SQLExecer result handler
// that assumes one row was affected in the passed query
func DefaultExecHandler(result sql.Result) (bool, error) {
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return affectedRows == int64(1), nil
}

// this makes sure the sql check is properly configured
func validateSQLConfig(cfg *SQLConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}

	if cfg.Execer == nil && cfg.Queryer == nil && cfg.Pinger == nil {
		return fmt.Errorf("one of Execer, Queryer, or Pinger is required in SQLConfig")
	}

	if (cfg.Execer != nil || cfg.Queryer != nil) && len(cfg.Query) == 0 {
		return fmt.Errorf("SQLConfig.Query is required")
	}

	return nil
}

// Status is used for performing a database ping against a dependency; it satisfies
// the "ICheckable" interface.
func (s *SQL) Status(ctx context.Context) (interface{}, error) {
	if err := validateSQLConfig(s.Config); err != nil {
		return nil, err
	}

	switch {
	// check for SQLExecer first
	case s.Config.Execer != nil:
		// if the result handler is nil, use the default
		if s.Config.ExecerResultHandler == nil {
			s.Config.ExecerResultHandler = DefaultExecHandler
		}
		// run the execer
		return s.runExecer()
	// check for SQLQueryer next
	case s.Config.Queryer != nil:
		// if the result handler is nil, use the default
		if s.Config.QueryerResultHandler == nil {
			s.Config.QueryerResultHandler = DefaultQueryHandler
		}
		// run the queryer
		return s.runQueryer()
	// finally, must be a pinger
	default:
		ctx := context.Background()
		return nil, s.Config.Pinger.PingContext(ctx)
	}
}

// This will run the execer from the Status func
func (s *SQL) runExecer() (interface{}, error) {
	ctx := context.Background()
	result, err := s.Config.Execer.ExecContext(ctx, s.Config.Query, s.Config.Params...)
	if err != nil {
		return nil, err
	}

	ok, err := s.Config.ExecerResultHandler(result)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("userland exec result handler returned false")
	}

	return nil, nil
}

// This will run the queryer from the Status func
func (s *SQL) runQueryer() (interface{}, error) {
	ctx := context.Background()
	rows, err := s.Config.Queryer.QueryContext(ctx, s.Config.Query, s.Config.Params...)
	if err != nil {
		return nil, err
	}

	// the BYO result handler is responsible for closing the rows

	ok, err := s.Config.QueryerResultHandler(rows)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("userland query result handler returned false")
	}

	return nil, nil
}
