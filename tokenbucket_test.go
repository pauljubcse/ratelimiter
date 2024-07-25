package ratelimiter

import (
	"fmt"
	"testing"

	"github.com/pauljubcse/kvs"
	"github.com/pauljubcse/kvsclient"
)


func TestTokenBucket(t *testing.T) {
	go kvs.StartServer("ws://localhost:9090/ws")
	// if err != nil {
	// 	log.Fatalf("Failed to start server: %v", err)
	// }
	fmt.Println("Started server...")

	client, err := kvsclient.NewClient("ws://localhost:9090/ws")
	
	if err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}
	defer client.Close()

	domain := "test_domain"
	client.CreateDomain(domain)
	interval := int64(20) // seconds
	maxRequests := int64(6)

	// Initialize the TokenBucket
	tb := &TokenBucket{
		dbclient:         client,
		domain:           domain,
		intervalInSeconds: interval,
		maximumRequests:  maxRequests,
	}

	// Define user IDs for testing
	userIDs := []string{
		"192.168.0.1", "192.168.0.2", "192.168.0.3",
		"192.168.0.4", "192.168.0.5", "192.168.0.6",
		"192.168.0.7", "192.168.0.8", "192.168.0.9", "192.168.0.10",
	}

	// Metrics
	successCount := 0
	dropCount := 0

	// Test rate limiting
	for i := int64(0); i < maxRequests+2; i++ {
		for _, userID := range userIDs {
			if tb.Allow(userID) {
				successCount++
			} else {
				dropCount++
			}
		}

		// Advance time to ensure the counter resets
		//time.Sleep(time.Duration(interval) * time.Second)
	}

	// Report metrics
	t.Logf("Success Count: %d", successCount)
	t.Logf("Drop Count: %d", dropCount)

	// Assertions
	expectedTotalRequests := int64(int64(len(userIDs)) * (maxRequests + 2))
	if int64(successCount+dropCount) != expectedTotalRequests {
		t.Errorf("Total requests should equal success + drop count, expected %d, got %d", expectedTotalRequests, successCount+dropCount)
	}

	if successCount == 0 {
		t.Errorf("There should be successful requests")
	}

	if dropCount == 0 {
		t.Errorf("There should be dropped requests")
	}
}