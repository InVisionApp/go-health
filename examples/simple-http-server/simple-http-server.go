package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/InVisionApp/go-health/v2"
	"github.com/InVisionApp/go-health/v2/checkers"
	"github.com/InVisionApp/go-health/v2/handlers"
)

func main() {
	// Create a new health instance
	h := health.New()
	goodTestURL, _ := url.Parse("https://google.com")
	badTestURL, _ := url.Parse("google.com")

	// Create a couple of checks
	goodHTTPCheck, _ := checkers.NewHTTP(&checkers.HTTPConfig{
		URL: goodTestURL,
	})

	badHTTPCheck, _ := checkers.NewHTTP(&checkers.HTTPConfig{
		URL: badTestURL,
	})

	// Add the checks to the health instance
	h.AddChecks([]*health.Config{
		{
			Name:     "good-check",
			Checker:  goodHTTPCheck,
			Interval: time.Duration(2) * time.Second,
			Fatal:    true,
		},
		{
			Name:     "bad-check",
			Checker:  badHTTPCheck,
			Interval: time.Duration(2) * time.Second,
			Fatal:    false,
		},
	})

	//  Start the healthcheck process
	if err := h.Start(); err != nil {
		log.Fatalf("Unable to start healthcheck: %v", err)
	}

	log.Println("Server listening on :8080")

	// Define a healthcheck endpoint and use the built-in JSON handler
	http.HandleFunc("/healthcheck", handlers.NewJSONHandlerFunc(h, nil))
	http.ListenAndServe(":8080", nil)
}
