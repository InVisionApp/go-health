package checkers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const execSQL = "UPDATE some_table"
const querySQL = "SELECT some_column"

type testHealthyPinger struct{}

type testUnhealthyPinger struct{}

type nilExecer struct{}

type nilQueryer struct{}

type fakeSQLResult struct{}

func (p *testHealthyPinger) PingContext(ctx context.Context) error {
	return nil
}

func (p *testUnhealthyPinger) PingContext(ctx context.Context) error {
	return fmt.Errorf("ping failed")
}

func (e *nilExecer) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (q *nilQueryer) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (r *fakeSQLResult) LastInsertId() (int64, error) {
	return 1, nil
}

func (r *fakeSQLResult) RowsAffected() (int64, error) {
	return 0, errors.New("affected rows failure")
}

func errExecHandler(result sql.Result) (bool, error) {
	return false, errors.New("exec handler failure")
}

func falseExecHandler(result sql.Result) (bool, error) {
	return false, nil
}

func errQueryHandler(rows *sql.Rows) (bool, error) {
	return false, errors.New("query handler failure")
}

func falseQueryHandler(rows *sql.Rows) (bool, error) {
	return false, nil
}

func TestValidateSQLConfig(t *testing.T) {
	RegisterTestingT(t)

	t.Run("happy path with pinger", func(t *testing.T) {
		hp := &testHealthyPinger{}
		cfg := &SQLConfig{
			Pinger: hp,
		}

		err := validateSQLConfig(cfg)
		Expect(err).To(BeNil())
	})

	t.Run("happy path with Execer", func(t *testing.T) {
		ex := &nilExecer{}
		cfg := &SQLConfig{
			Execer: ex,
			Query:  "not important",
		}

		err := validateSQLConfig(cfg)
		Expect(err).To(BeNil())
	})

	t.Run("sad path with nil config", func(t *testing.T) {
		err := validateSQLConfig(nil)
		Expect(err).ToNot(BeNil())
	})

	t.Run("sad path with Queryer and no query", func(t *testing.T) {
		q := &nilQueryer{}
		cfg := &SQLConfig{
			Queryer: q,
		}

		err := validateSQLConfig(cfg)
		Expect(err).ToNot(BeNil())
	})

	t.Run("sad path with Execer and no query", func(t *testing.T) {
		ex := &nilExecer{}
		cfg := &SQLConfig{
			Execer: ex,
		}

		err := validateSQLConfig(cfg)
		Expect(err).ToNot(BeNil())
	})

	t.Run("sad path with no actor", func(t *testing.T) {
		cfg := &SQLConfig{}

		err := validateSQLConfig(cfg)
		Expect(err).ToNot(BeNil())
	})
}

func TestNewSQL(t *testing.T) {
	RegisterTestingT(t)

	t.Run("happy path with db mock", func(t *testing.T) {
		db, _, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		s, err := NewSQL(&SQLConfig{
			Pinger: db,
		})
		Expect(err).To(BeNil())
		Expect(s).ToNot(BeNil())
	})

	t.Run("sad path when config is nil", func(t *testing.T) {
		s, err := NewSQL(nil)
		Expect(err).ToNot(BeNil())
		Expect(s).To(BeNil())
	})
}

func TestSQLStatus(t *testing.T) {
	RegisterTestingT(t)

	t.Run("happy path with db mock", func(t *testing.T) {
		db, _, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		s, err := NewSQL(&SQLConfig{
			Pinger: db,
		})
		Expect(err).To(BeNil())
		Expect(s).ToNot(BeNil())

		nothing, err := s.Status()
		Expect(err).ToNot(HaveOccurred())

		// status check returns no artifacts
		Expect(nothing).To(BeNil())
	})

	t.Run("SQLPinger returns healthy", func(t *testing.T) {
		db := &testHealthyPinger{}
		s, err := NewSQL(&SQLConfig{
			Pinger: db,
		})
		Expect(err).To(BeNil())
		Expect(s).ToNot(BeNil())

		_, err = s.Status()
		Expect(err).ToNot(HaveOccurred())
	})

	t.Run("IPingable returns unhealthy", func(t *testing.T) {
		db := &testUnhealthyPinger{}
		s, err := NewSQL(&SQLConfig{
			Pinger: db,
		})
		Expect(err).To(BeNil())
		Expect(s).ToNot(BeNil())

		_, err = s.Status()
		Expect(err).To(HaveOccurred())
	})

	t.Run("bad config", func(t *testing.T) {
		s := &SQL{}
		_, err := s.Status()
		Expect(err).To(HaveOccurred())
	})
}

func TestDefaultExecHandler(t *testing.T) {
	RegisterTestingT(t)

	t.Run("happy path", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		mock.ExpectExec(execSQL).WillReturnResult(sqlmock.NewResult(1, 1))
		s, err := NewSQL(&SQLConfig{
			Execer: db,
			Query:  execSQL,
		})
		Expect(err).To(BeNil())

		_, err = s.Status()
		Expect(err).To(BeNil())

	})

	t.Run("direct call with failing Result interface", func(t *testing.T) {
		result := &fakeSQLResult{}
		ok, err := DefaultExecHandler(result)

		Expect(err).ToNot(BeNil())
		Expect(ok).To(BeFalse())
	})
}

func TestRunExecer(t *testing.T) {
	RegisterTestingT(t)

	t.Run("ExecContext fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		mock.ExpectExec(execSQL).WillReturnError(errors.New("exec error"))

		s, err := NewSQL(&SQLConfig{
			Execer: db,
			Query:  execSQL,
		})
		Expect(err).To(BeNil())

		_, err = s.runExecer()
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("exec error"))
	})

	t.Run("ExecerResultHandler fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		mock.ExpectExec(execSQL).WillReturnResult(sqlmock.NewResult(1, 1))

		s, err := NewSQL(&SQLConfig{
			Execer:              db,
			Query:               execSQL,
			ExecerResultHandler: errExecHandler,
		})
		Expect(err).To(BeNil())

		_, err = s.runExecer()
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("exec handler failure"))
	})

	t.Run("ExecerResultHandler returns false", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		mock.ExpectExec(execSQL).WillReturnResult(sqlmock.NewResult(1, 1))

		s, err := NewSQL(&SQLConfig{
			Execer:              db,
			Query:               execSQL,
			ExecerResultHandler: falseExecHandler,
		})
		Expect(err).To(BeNil())

		_, err = s.runExecer()
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("userland exec result handler returned false"))
	})
}

func TestDefaultQueryHandler(t *testing.T) {
	RegisterTestingT(t)

	t.Run("happy path", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		rows := &sqlmock.Rows{}
		rows.AddRow()
		mock.ExpectQuery(querySQL).WillReturnRows(rows)
		s, err := NewSQL(&SQLConfig{
			Queryer: db,
			Query:   querySQL,
		})
		Expect(err).To(BeNil())

		_, err = s.Status()
		Expect(err).To(BeNil())

	})
}

func TestRunQueryer(t *testing.T) {
	RegisterTestingT(t)

	t.Run("QueryContext fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		mock.ExpectQuery(querySQL).WillReturnError(errors.New("query error"))

		s, err := NewSQL(&SQLConfig{
			Queryer: db,
			Query:   querySQL,
		})
		Expect(err).To(BeNil())

		_, err = s.runQueryer()
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("query error"))
	})

	t.Run("QueryResultHandler fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		rows := &sqlmock.Rows{}
		rows.AddRow()
		mock.ExpectQuery(querySQL).WillReturnRows(rows)

		s, err := NewSQL(&SQLConfig{
			Queryer:              db,
			Query:                querySQL,
			QueryerResultHandler: errQueryHandler,
		})
		Expect(err).To(BeNil())

		_, err = s.runQueryer()
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("query handler failure"))
	})

	t.Run("QueryerResultHandler returns false", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		rows := &sqlmock.Rows{}
		rows.AddRow()
		mock.ExpectQuery(querySQL).WillReturnRows(rows)

		s, err := NewSQL(&SQLConfig{
			Queryer:              db,
			Query:                querySQL,
			QueryerResultHandler: falseQueryHandler,
		})
		Expect(err).To(BeNil())

		_, err = s.runQueryer()
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("userland query result handler returned false"))
	})
}
