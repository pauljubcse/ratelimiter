package ratelimiter

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pauljubcse/kvsclient"
)

type SlidingWindowRateLimiter struct {
	dbclient          *kvsclient.Client
	domain            string
	intervalInSeconds int64
	maximumRequests   int64
}

// RateLimitUsingSlidingWindow implements the sliding window rate limiter
func (s *SlidingWindowRateLimiter) Allow(userID string) bool {
	now := time.Now().Unix()

	// Define the current and last time windows
	currentWindow := strconv.FormatInt(now/s.intervalInSeconds, 10)
	lastWindow := strconv.FormatInt((now-s.intervalInSeconds)/s.intervalInSeconds, 10)

	// Define keys for storing request counts in the current and last windows
	currentKey := userID + ":" + currentWindow
	lastKey := userID + ":" + lastWindow

	// Get current window request count
	value, err := s.dbclient.GetString(s.domain, currentKey)
	if err != nil {
		fmt.Printf("Failed to get current window count: %v\n", err)
		err1 := s.dbclient.SetString(s.domain, currentKey, strconv.FormatInt(0, 10))
		if err1 != nil{
			return false
		}
		value="0"
	}
	requestCountCurrentWindow, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		fmt.Printf("Failed to parse current window count: %v\n", err)
		return false
	}

	// Get last window request count
	value, err = s.dbclient.GetString(s.domain, lastKey)
	if err != nil {
		fmt.Printf("Failed to get last window count: %v\n", err)
		err1 := s.dbclient.SetString(s.domain, lastKey, strconv.FormatInt(0, 10))
		if err1 != nil{
			return false
		}
		value="0"
	}
	requestCountLastWindow, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		fmt.Printf("Failed to parse last window count: %v\n", err)
		return false
	}

	elapsedTimePercentage := float64(now%int64(s.intervalInSeconds)) / float64(s.intervalInSeconds)

	// Calculate the weighted count of the last window plus the current window count
	weightedCountLastWindow := float64(requestCountLastWindow) * (1 - elapsedTimePercentage)
	totalRequestCount := weightedCountLastWindow + float64(requestCountCurrentWindow)

	if totalRequestCount >= float64(s.maximumRequests) {
		// Drop request if the limit is exceeded
		return false
	}

	// Increment request count by 1 in the current window
	err = s.dbclient.Increment(s.domain, currentKey)
	if err != nil {
		fmt.Printf("Failed to increment current window count: %v\n", err)
		return false
	}

	// Handle request
	return true
}
