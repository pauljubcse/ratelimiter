package ratelimiter

import (
	"fmt"
	"testing"
	"time"

	"github.com/pauljubcse/kvs"
	"github.com/pauljubcse/kvsclient"
)

func TestFixedWindowLimiter(t *testing.T) {
	go kvs.StartServer("ws://localhost:9090/ws")
	fmt.Println("Started server...")

	client, err := kvsclient.NewClient("ws://localhost:9090/ws")
	if err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}
	defer client.Close()

	domain := "test_domain"
	client.CreateDomain(domain)

	limiter := FixedWindowLimiter{
		dbclient:          client,
		domain:            domain,
		intervalInSeconds: 3,  // 10-second window
		maximumRequests:   5,   // Allow 5 requests per window
	}

	userIDs := []string{
		"192.168.0.1", "192.168.0.2", "192.168.0.3",
		"192.168.0.4", "192.168.0.5", "192.168.0.6",
		"192.168.0.7", "192.168.0.8", "192.168.0.9", "192.168.0.10",
	}
	successCount := 0
	dropCount := 0
	iterations := 3
	for i:=0;i<iterations;i++{
		for _, userID := range userIDs {
			fmt.Printf("Testing requests for userID: %s\n", userID)
			//fmt.Printf("Test: %s\n", client.GetString(domain, userID+":"+))
			for i := 0; i < 6; i++ { // Attempt 6 requests per user
				if limiter.Allow(userID) {
					fmt.Printf("Request %d for userID %s allowed\n", i+1, userID)
					successCount++
				} else {
					fmt.Printf("Request %d for userID %s dropped\n", i+1, userID)
					dropCount++
				}
			}
		}
		time.Sleep(time.Second*5)
	}

	fmt.Printf("Total successful requests: %d\n", successCount)
	fmt.Printf("Total dropped requests: %d\n", dropCount)

	// Check if the correct number of requests were allowed and dropped
	expectedSuccessCount := iterations*10 * 5 // 10 users * 5 allowed requests per user
	expectedDropCount := iterations*10 * 1    // 10 users * 1 dropped request per user

	if successCount != expectedSuccessCount {
		t.Errorf("Expected %d successful requests, got %d", expectedSuccessCount, successCount)
	}
	if dropCount != expectedDropCount {
		t.Errorf("Expected %d dropped requests, got %d", expectedDropCount, dropCount)
	}
}
