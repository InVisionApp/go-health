package checkers

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type testInvalidKind1 struct{}
type testInvalidKind2 struct{}
type testInvalidKind3 struct{}
type testInvalidKind4 struct{}
type testHealthyIPingable struct{}
type testUnhealthyIPingable struct{}
type testHealthyPinger struct{}
type testUnhealthyPinger struct{}

func (iv *testInvalidKind1) Ping()                    {}
func (iv *testInvalidKind2) Ping(ctx context.Context) {}
func (iv *testInvalidKind3) Ping() (int, error) {
	return 0, nil
}
func (iv *testInvalidKind4) Ping(ctx context.Context) (int, error) {
	return 0, nil
}
func (p *testHealthyIPingable) Ping() error {
	return nil
}
func (p *testUnhealthyIPingable) Ping() error {
	return fmt.Errorf("ping failed")
}
func (p *testHealthyPinger) Ping(ctx context.Context) error {
	return nil
}
func (p *testUnhealthyPinger) Ping(ctx context.Context) error {
	return fmt.Errorf("ping failed")
}

func TestNewSQL(t *testing.T) {
	RegisterTestingT(t)

	t.Run("happy path with db mock", func(t *testing.T) {
		db, _, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		s, err := NewSQL(&SQLConfig{
			DB: db,
		})
		Expect(err).To(BeNil())
		Expect(s).ToNot(BeNil())
	})

	t.Run("sad path when config is nil", func(t *testing.T) {
		_, err := NewSQL(nil)
		Expect(err).ToNot(BeNil())
	})

	t.Run("sad path when DB is nil", func(t *testing.T) {
		_, err := NewSQL(&SQLConfig{
			DB: nil,
		})
		Expect(err).ToNot(BeNil())
	})

	t.Run("sad path using invalid interfaces", func(t *testing.T) {
		var err error

		// testInvalidKind1 does not implement IPingable
		// because it does not return an error
		iv1 := &testInvalidKind1{}
		_, err = NewSQL(&SQLConfig{
			DB: iv1,
		})
		Expect(err).ToNot(BeNil())

		// testInvalidKind2 does not implement Pinger
		// because it does not return an error
		iv2 := &testInvalidKind2{}
		_, err = NewSQL(&SQLConfig{
			DB: iv2,
		})
		Expect(err).ToNot(BeNil())

		// testInvalidKind3 does not implement IPingable
		// because it returns multiple values
		iv3 := &testInvalidKind1{}
		_, err = NewSQL(&SQLConfig{
			DB: iv3,
		})
		Expect(err).ToNot(BeNil())

		// testInvalidKind4 does not implement Pinger
		// because it returns multiple values
		iv4 := &testInvalidKind1{}
		_, err = NewSQL(&SQLConfig{
			DB: iv4,
		})
		Expect(err).ToNot(BeNil())
	})
}

func TestSQLStatus(t *testing.T) {
	RegisterTestingT(t)

	t.Run("happy path with db mock", func(t *testing.T) {
		db, _, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		s, err := NewSQL(&SQLConfig{
			DB: db,
		})
		Expect(err).To(BeNil())
		Expect(s).ToNot(BeNil())

		nothing, err := s.Status()
		Expect(err).ToNot(HaveOccurred())

		// status check returns no artifacts
		Expect(nothing).To(BeNil())
	})

	t.Run("IPingable returns healthy", func(t *testing.T) {
		db := &testHealthyIPingable{}
		s, err := NewSQL(&SQLConfig{
			DB: db,
		})
		Expect(err).To(BeNil())
		Expect(s).ToNot(BeNil())

		_, err = s.Status()
		Expect(err).ToNot(HaveOccurred())
	})

	t.Run("IPingable returns unhealthy", func(t *testing.T) {
		db := &testUnhealthyIPingable{}
		s, err := NewSQL(&SQLConfig{
			DB: db,
		})
		Expect(err).To(BeNil())
		Expect(s).ToNot(BeNil())

		_, err = s.Status()
		Expect(err).To(HaveOccurred())
	})

	t.Run("Pinger returns healthy", func(t *testing.T) {
		db := &testHealthyPinger{}
		s, err := NewSQL(&SQLConfig{
			DB: db,
		})
		Expect(err).To(BeNil())
		Expect(s).ToNot(BeNil())

		_, err = s.Status()
		Expect(err).ToNot(HaveOccurred())
	})

	t.Run("IPingable returns unhealthy", func(t *testing.T) {
		db := &testUnhealthyPinger{}
		s, err := NewSQL(&SQLConfig{
			DB: db,
		})
		Expect(err).To(BeNil())
		Expect(s).ToNot(BeNil())

		_, err = s.Status()
		Expect(err).To(HaveOccurred())
	})
}
