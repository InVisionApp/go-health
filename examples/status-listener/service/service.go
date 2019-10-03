package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/InVisionApp/go-health/v2"
	"github.com/InVisionApp/go-health/v2/checkers"
	"github.com/InVisionApp/go-health/v2/handlers"
)

var svcLogger *log.Logger

// HealthCheckStatusListener is the implementation of the IStatusListener interface
type HealthCheckStatusListener struct{}

// HealthCheckFailed is triggered when a health check fails the first time
func (sl *HealthCheckStatusListener) HealthCheckFailed(entry *health.State) {
	svcLogger.Printf("State for failed health check: %+v", entry)
}

// HealthCheckRecovered is triggered when a health check recovers
func (sl *HealthCheckStatusListener) HealthCheckRecovered(entry *health.State, recordedFailures int64, failureDurationSeconds float64) {
	svcLogger.Printf("Recovering from %d contiguous errors, lasting %1.2f seconds: %+v", recordedFailures, failureDurationSeconds, entry)
}

func init() {
	svcLogger = log.New(os.Stderr, "service: ", 0)
}

func main() {
	// Create a new health instance
	h := health.New()
	// disable logging from health lib
	h.DisableLogging()
	testURL, _ := url.Parse("http://0.0.0.0:8081")

	// Create a couple of checks
	httpCheck, _ := checkers.NewHTTP(&checkers.HTTPConfig{
		URL: testURL,
	})

	// Add the checks to the health instance
	h.AddChecks([]*health.Config{
		{
			Name:     "dependency-check",
			Checker:  httpCheck,
			Interval: time.Duration(2) * time.Second,
			Fatal:    true,
		},
	})

	// set status listener
	sl := &HealthCheckStatusListener{}
	h.StatusListener = sl

	//  Start the healthcheck process
	if err := h.Start(); err != nil {
		svcLogger.Fatalf("Unable to start healthcheck: %v", err)
	}

	svcLogger.Println("Server listening on :8080")

	// Define a healthcheck endpoint and use the built-in JSON handler
	http.HandleFunc("/healthcheck", handlers.NewJSONHandlerFunc(h, nil))
	http.ListenAndServe(":8080", nil)
}
