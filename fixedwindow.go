package ratelimiter

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pauljubcse/kvsclient"
)

type FixedWindowLimiter struct {
	dbclient          *kvsclient.Client
	domain            string
	intervalInSeconds int64
	maximumRequests   int64
}

func (f *FixedWindowLimiter) Allow(userID string) bool {
	// Calculate the current window
	currentWindow := strconv.FormatInt(time.Now().Unix()/f.intervalInSeconds, 10)
	key := userID + ":" + currentWindow // userID + current time window

	// Retrieve the current window count from KVS
	value, err := f.dbclient.GetString(f.domain, key)
	if err != nil {
		// Handle error (e.g., key not found)
		value = "0"
		f.dbclient.SetString(f.domain, key, value)
	}
	requestCount, _ := strconv.ParseInt(value, 10, 64)
	fmt.Printf("Current request count for %s: %d\n", key, requestCount)

	if requestCount >= f.maximumRequests {
		// Drop request if the limit is reached
		fmt.Println("Request limit reached, dropping request.")
		return false
	}

	// Increment the request count
	err = f.dbclient.Increment(f.domain, key)
	if err != nil {
		// Handle error (e.g., failed to increment value)
		fmt.Printf("Error incrementing request count: %v\n", err)
		return false
	}
	fmt.Println("Request allowed.")
	return true
}
