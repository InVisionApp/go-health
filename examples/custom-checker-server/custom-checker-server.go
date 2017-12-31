package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/handlers"
)

type CustomCheck struct{}

func main() {
	// Create a new health instance
	h := health.New()

	// Instantiate your custom check
	customCheck := &CustomCheck{}

	// Add the checks to the health instance
	h.AddChecks([]*health.Config{
		{
			Name:     "good-check",
			Checker:  customCheck,
			Interval: time.Duration(2) * time.Second,
			Fatal:    true,
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

// Satisfy the go-health.ICheckable interface
func (c *CustomCheck) Status() (interface{}, error) {
	// perform some sort of check
	if false {
		return nil, fmt.Errorf("Something major just broke")
	}

	// You can return additional information pertaining to the check as long
	// as it can be JSON marshalled
	return map[string]int{"foo": 123, "bar": 456}, nil
}
