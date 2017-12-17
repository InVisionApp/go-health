package checkers

import (
	"net/http"
	"net/http/httptest"
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

func TestHTTPStatus(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		testURL, err := url.Parse(ts.URL)
		Expect(err).ToNot(HaveOccurred())

		cfg := &HTTPConfig{
			URL: testURL,
		}

		checker, err := NewHTTP(cfg)
		Expect(err).ToNot(HaveOccurred())

		data, err := checker.Status()
		Expect(err).ToNot(HaveOccurred())
		Expect(data).To(BeNil())
	})

	t.Run("Should return error if HTTP call fails", func(t *testing.T) {
		testURL, _ := url.Parse("no-scheme.com")

		cfg := &HTTPConfig{
			URL: testURL,
		}

		checker, err := NewHTTP(cfg)
		Expect(err).ToNot(HaveOccurred())

		data, err := checker.Status()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported protocol"))
		Expect(data).To(BeNil())
	})

	t.Run("Should return error if expected response status code does not match received status code", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		testURL, err := url.Parse(ts.URL)
		Expect(err).ToNot(HaveOccurred())

		cfg := &HTTPConfig{
			URL: testURL,
		}

		checker, err := NewHTTP(cfg)
		Expect(err).ToNot(HaveOccurred())

		data, err := checker.Status()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("does not match expected status code"))
		Expect(data).To(BeNil())
	})

	t.Run("Should return error if response data does not contain expected data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo"))
		}))
		defer ts.Close()

		testURL, err := url.Parse(ts.URL)
		Expect(err).ToNot(HaveOccurred())

		cfg := &HTTPConfig{
			URL:    testURL,
			Expect: "bar",
		}

		checker, err := NewHTTP(cfg)
		Expect(err).ToNot(HaveOccurred())

		data, err := checker.Status()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("does not contain expected content"))
		Expect(data).To(BeNil())
	})

	t.Run("Should not error if expected data in response is found", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo"))
		}))
		defer ts.Close()

		testURL, err := url.Parse(ts.URL)
		Expect(err).ToNot(HaveOccurred())

		cfg := &HTTPConfig{
			URL:    testURL,
			Expect: "foo",
		}

		checker, err := NewHTTP(cfg)
		Expect(err).ToNot(HaveOccurred())

		data, err := checker.Status()
		Expect(err).ToNot(HaveOccurred())
		Expect(data).To(BeNil())
	})

	t.Run("Should return error if response body is not readable", func(t *testing.T) {

	})
}
