package checkers

import (
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"fmt"

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

func TestDo(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error with unparsable payload", func(t *testing.T) {
		h := &HTTP{
			Config: &HTTPConfig{
				Payload: math.NaN(),
			},
		}

		res, err := h.do()
		Expect(err).To(HaveOccurred())
		Expect(res).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("error parsing payload"))
		res.Close()
	})

	t.Run("Should error if request can't be created", func(t *testing.T) {
		u, _ := url.Parse("http://google.com")
		h := &HTTP{
			Config: &HTTPConfig{
				Payload: "foo",
				Method:  "bad method",
				URL:     u,
			},
		}

		res, err := h.do()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unable to create new HTTP request for HTTPMonitor check"))
		Expect(res).To(BeNil())
		res.Close()
	})
}

func TestPrepare(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error when URL is not set", func(t *testing.T) {
		h := &HTTPConfig{}
		err := h.prepare()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("URL cannot be nil"))
	})

	t.Run("Should set appropriate defaults", func(t *testing.T) {
		u, _ := url.Parse("http://google.com")
		h := &HTTPConfig{URL: u}

		err := h.prepare()
		Expect(err).ToNot(HaveOccurred())

		Expect(h.StatusCode).To(Equal(http.StatusOK))
		Expect(h.Method).To(Equal("GET"))
		Expect(h.Timeout).To(Equal(defaultHTTPTimeout))
		Expect(h.Client.Timeout).To(Equal(h.Timeout))
	})

	t.Run("Custom http client timeout should be updated", func(t *testing.T) {
		u, _ := url.Parse("http://google.com")
		h := &HTTPConfig{URL: u, Client: &http.Client{}, Timeout: time.Duration(1) * time.Second}

		err := h.prepare()
		Expect(err).ToNot(HaveOccurred())
		Expect(h.Client.Timeout).To(Equal(time.Duration(1) * time.Second))
	})
}

func TestParsePayload(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		reader, err := parsePayload("test")
		Expect(err).ToNot(HaveOccurred())
		Expect(reader).ToNot(BeNil())

		reader, err = parsePayload([]byte("test"))
		Expect(err).ToNot(HaveOccurred())
		Expect(reader).ToNot(BeNil())

		foo := struct {
			Name string
		}{
			Name: "foo",
		}

		reader, err = parsePayload(foo)
		Expect(err).ToNot(HaveOccurred())
		Expect(reader).ToNot(BeNil())
	})

	t.Run("should return nil if no payload is passed in", func(t *testing.T) {
		reader, err := parsePayload(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(reader).To(BeNil())
	})

	t.Run("should error with payload that can't be marshalled to json", func(t *testing.T) {
		reader, err := parsePayload(math.NaN())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to marshal json body"))
		Expect(reader).To(BeNil())
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
		httpClient := &http.Client{
			Transport: newTransport(),
		}

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
			Client: httpClient,
		}

		checker, err := NewHTTP(cfg)
		Expect(err).ToNot(HaveOccurred())

		data, err := checker.Status()

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unable to read response body to perform content expectancy check"))
		Expect(data).To(BeNil())
	})
}

type CustomTransport struct{}

func newTransport() *CustomTransport {
	return &CustomTransport{}
}

func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       &mockReader{},
	}, nil
}

type mockReader struct{}

func (m *mockReader) Read(p []byte) (n int, err error) { return 0, fmt.Errorf("foo") }
func (m *mockReader) Close() error                     { return nil }
