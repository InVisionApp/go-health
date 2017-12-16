package checkers

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewRedis(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {

	})

	t.Run("Bad config should error", func(t *testing.T) {

	})
}

func TestValidateRedisConfig(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error with no client", func(t *testing.T) {

	})

	t.Run("Should error if none of the check methods are enabled", func(t *testing.T) {

	})

	t.Run("Should error if .Set is used but key is undefined", func(t *testing.T) {

	})

	t.Run("Should error if .Get is used but key is undefined", func(t *testing.T) {

	})
}

func TestRedisStatus(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error when ping is enabled and fails", func(t *testing.T) {

	})

	t.Run("When set is enabled", func(t *testing.T) {
		t.Run("should error if set fails", func(t *testing.T) {

		})

		t.Run("should use .Value if .Value is defined", func(t *testing.T) {

		})
	})

	t.Run("When get is enabled", func(t *testing.T) {
		t.Run("should error if key is missing and NoErrorMissingKey not set", func(t *testing.T) {

		})

		t.Run("should NOT error if key is missing and NoErrorMissingKey IS set", func(t *testing.T) {

		})

		t.Run("should error if get fails", func(t *testing.T) {

		})

		t.Run("should error if .Expect is set and the value does not match", func(t *testing.T) {

		})

		t.Run("should NOT error if .Expect is not set", func(t *testing.T) {

		})
	})
}
