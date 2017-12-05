package checkers

import (
	"net/url"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewHTTP(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		u, err := url.Parse("http://testing.com/search?q=foo")
		Expect(err).ToNot(HaveOccurred())

		h, err := NewHTTP(&HTTPConfig{
			URL: u,
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(h).ToNot(BeNil())
	})

	t.Run("Should error with a nil cfg", func(t *testing.T) {
		h, err := NewHTTP(nil)

		Expect(h).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("Passed in config cannot be nil"))
	})

	t.Run("Should error when prepare fails", func(t *testing.T) {
		h, err := NewHTTP(&HTTPConfig{})

		Expect(h).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("URL cannot be nil"))
	})
}
