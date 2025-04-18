package ratelimiter

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/pauljubcse/kvs"
	"github.com/pauljubcse/kvsclient"
)

//Issue occurs when two requests occur at same nanosecond
func TestSlidingWindowLogsLimiter(t *testing.T) {
	go kvs.StartServer("ws://localhost:9190/ws")

	client, err := kvsclient.NewClient("ws://localhost:9190/ws")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()

	// Create a domain
	domain := "test_domain"
	err = client.CreateDomain(domain)
	if err != nil {
		t.Fatalf("Failed to create domain: %v", err)
	}

	// Create a limiter instance
	limiter := &SlidingWindowLogsLimiter{
		dbclient:          client,
		domain:            domain,
		intervalInSeconds: 5, // 1 minute
		maximumRequests:   5,
	}

	// Test cases
	tests := []struct {
		name       string
		userID     string
		uniqueID   string
		expectPass bool
	}{
		{"First Request", "user1", "req1", true},
		{"Second Request", "user1", "req2", true},
		{"Third Request", "user1", "req3", true},
		{"Fourth Request", "user1", "req4", true},
		{"Fifth Request", "user1", "req5", true},
		{"Sixth Request - Should Fail", "user1", "req6", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := limiter.Allow(tt.userID)
			if result != tt.expectPass {
				t.Errorf("Expected %v but got %v", tt.expectPass, result)
			}
			time.Sleep(time.Nanosecond*1)
			// Print rank difference
			if tt.name == "Sixth Request - Should Fail"{
				currentTime := time.Now().Unix()
				lastWindowTime := currentTime - limiter.intervalInSeconds
				slkey := tt.userID + ":sl"

				r1, err := client.RankInSkipList(limiter.domain, slkey, strconv.FormatInt(lastWindowTime, 10))
				if err != nil {
					t.Fatalf("Failed to get rank for last window time: %v", err)
				}
				rankLastWindow, err := strconv.ParseInt(r1, 10, 64)
				if err != nil {
					t.Fatalf("Failed to parse rank for last window time: %v", err)
				}

				r2, err := client.RankInSkipList(limiter.domain, slkey, strconv.FormatInt(currentTime, 10))
				if err != nil {
					t.Fatalf("Failed to get rank for current time: %v", err)
				}
				rankCurrentTime, err := strconv.ParseInt(r2, 10, 64)
				if err != nil {
					t.Fatalf("Failed to parse rank for current time: %v", err)
				}

				fmt.Printf("Rank difference for failed request: %d\n", rankCurrentTime-rankLastWindow)
			}
		})
	}
}

func TestSlidingWindowLogsLimiterWithTimeIntervals(t *testing.T) {
	go kvs.StartServer("ws://localhost:9190/ws")

	client, err := kvsclient.NewClient("ws://localhost:9190/ws")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()

	// Create a domain
	domain := "test_domain_interval"
	err = client.CreateDomain(domain)
	if err != nil {
		t.Fatalf("Failed to create domain: %v", err)
	}

	// Create a limiter instance
	limiter := &SlidingWindowLogsLimiter{
		dbclient:          client,
		domain:            domain,
		intervalInSeconds: 5,
		maximumRequests:   3,
	}

	// Test cases
	tests := []struct {
		name       string
		userID     string
		//uniqueID   string
		sleepTime  time.Duration
		expectPass bool
	}{
		{"First Request", "user1", 0, true},
		{"Second Request", "user1", 2 * time.Second, true},
		//{"Third Request - Should Fail", "user1", 0, false},
		{"Sleep 2 seconds", "user1", 2 * time.Second, true},
		{"Fourth Request", "user1", 0, false},
		//{"Fifth Request - Should Fail", "user1", "req5", 0, false},
		{"Sleep 2 seconds", "user1", 2 * time.Second, false},
		{"Sixth Request", "user1", 0, true},
		{"Seventh Request", "user1", 0, true},
		{"Eigth Request", "user1", 0, false},
	}
	then:=time.Now().UnixNano()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			
			result := limiter.Allow(tt.userID)
			if result != tt.expectPass {
				t.Errorf("Expected %v but got %v", tt.expectPass, result)
			}
			if tt.sleepTime > 0 {
				time.Sleep(tt.sleepTime)
			}
			time.Sleep(time.Nanosecond*1)
			fmt.Printf("Time: %d\n", (time.Now().UnixNano()-then))
		})
	}
}