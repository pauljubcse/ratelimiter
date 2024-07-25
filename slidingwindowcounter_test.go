package ratelimiter

import (
	"testing"
	"time"

	"github.com/pauljubcse/kvs"
	"github.com/pauljubcse/kvsclient"
	//"github.com/stretchr/testify/assert"
)

func TestRateLimitUsingSlidingWindow(t *testing.T) {
	// Setup
	go kvs.StartServer("ws://localhost:9000/ws")
	client, err := kvsclient.NewClient("ws://localhost:9000/ws")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()

	domain := "test_domain"
	err = client.CreateDomain(domain)
	if err != nil {
		t.Fatalf("Failed to create domain: %v", err)
	}

	// Configure the rate limiter
	rateLimiter := &SlidingWindowRateLimiter{
		dbclient:          client,
		domain:            domain,
		intervalInSeconds: 10, // 1 minute window
		maximumRequests:   5,
	}

	// Test cases
	tests := []struct {
		name           string
		requests       int
		expectedResult bool
	}{
		{"Allow within limit", 3, true},   // Should allow the request
		{"Allow at limit", 2, true},       // Should allow the request
		{"Deny above limit", 3, false},    // Should deny the request
		{"Reset after interval", 1, true}, // After waiting for interval, should allow again
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Reset after interval" {
				time.Sleep(time.Duration(rateLimiter.intervalInSeconds+1) * time.Second) // Wait for the interval to reset
			}

			// Insert or update the log entries based on the test case
			for i := 0; i < tt.requests; i++ {
				result := rateLimiter.Allow("user1")
				if !tt.expectedResult && result {
					t.Errorf("Expected request to be denied, but it was allowed")
				}
				if tt.expectedResult && !result {
					t.Errorf("Expected request to be allowed, but it was denied")
				}
			}

			// Optionally: Check the internal state or logs in kvsclient if needed
			// e.g., verifying the counts in the skip list (not shown here)
		})
	}
}
