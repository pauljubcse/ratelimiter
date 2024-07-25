package ratelimiter

import (
	//"context"

	"strconv"
	"time"

	"github.com/pauljubcse/kvsclient"
)

//Lazy Refill
type TokenBucket struct{
	dbclient *kvsclient.Client
	domain string
	intervalInSeconds int64
	maximumRequests int64

}
func (t *TokenBucket) Allow (userID string) bool {

	// Retrieve the last reset time from KVS
	lastResetTimeStr, err := t.dbclient.GetString(t.domain, userID+"_last_reset_time")
	if err != nil {
		// Handle error (e.g., key not found or context cancellation)
		lastResetTimeStr = "0"
	}

	lastResetTime, _ := strconv.ParseInt(lastResetTimeStr, 10, 64)

	// Check if the time window has elapsed
	if time.Now().Unix()-lastResetTime >= t.intervalInSeconds {
		// Reset the counter and last reset time
		err := t.dbclient.SetString(t.domain, userID+"_counter", strconv.FormatInt(t.maximumRequests, 10))
		if err != nil {
			// Handle error (e.g., failed to set value)
			return false
		}
		err = t.dbclient.SetString(t.domain, userID+"_last_reset_time", strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
			// Handle error (e.g., failed to set value)
			return false
		}
	} else {
		// Retrieve the current counter value
		counterStr, err := t.dbclient.GetString(t.domain, userID+"_counter")
		if err != nil {
			// Handle error (e.g., key not found or context cancellation)
			counterStr = strconv.FormatInt(t.maximumRequests, 10)
		}

		requestsLeft, _ := strconv.ParseInt(counterStr, 10, 64)

		if requestsLeft <= 0 {
			// Drop request if no requests are left
			return false
		}
	}

	// Decrement the request count
	err = t.dbclient.Decrement(t.domain, userID+"_counter")
	return err==nil
}

// func (t *TokenBucket) Allow(userID string) bool {
// 	fmt.Printf("Checking rate limit for user: %s\n", userID)

// 	// Retrieve the last reset time from KVS
// 	lastResetTimeStr, err := t.dbclient.GetString(t.domain, userID+"_last_reset_time")
// 	if err != nil {
// 		// Handle error (e.g., key not found or context cancellation)
// 		fmt.Printf("Error retrieving last reset time: %v\n", err)
// 		lastResetTimeStr = "0"
// 	}
// 	fmt.Printf("Last reset time: %s\n", lastResetTimeStr)

// 	lastResetTime, _ := strconv.ParseInt(lastResetTimeStr, 10, 64)

// 	// Check if the time window has elapsed
// 	if time.Now().Unix()-lastResetTime >= t.intervalInSeconds {
// 		// Reset the counter and last reset time
// 		fmt.Println("Time window has elapsed, resetting counter and last reset time.")
// 		err := t.dbclient.SetString(t.domain, userID+"_counter", strconv.FormatInt(t.maximumRequests, 10))
// 		if err != nil {
// 			// Handle error (e.g., failed to set value)
// 			fmt.Printf("Error setting counter: %v\n", err)
// 			return false
// 		}
// 		err = t.dbclient.SetString(t.domain, userID+"_last_reset_time", strconv.FormatInt(time.Now().Unix(), 10))
// 		if err != nil {
// 			// Handle error (e.g., failed to set value)
// 			fmt.Printf("Error setting last reset time: %v\n", err)
// 			return false
// 		}
// 	} else {
// 		// Retrieve the current counter value
// 		counterStr, err := t.dbclient.GetString(t.domain, userID+"_counter")
// 		if err != nil {
// 			// Handle error (e.g., key not found or context cancellation)
// 			fmt.Printf("Error retrieving counter: %v\n", err)
// 			counterStr = strconv.FormatInt(t.maximumRequests, 10)
// 		}
// 		fmt.Printf("Current counter value: %s\n", counterStr)

// 		requestsLeft, _ := strconv.ParseInt(counterStr, 10, 64)

// 		if requestsLeft <= 0 {
// 			// Drop request if no requests are left
// 			fmt.Println("Request limit reached, dropping request.")
// 			return false
// 		}
// 	}

// 	// Decrement the request count
// 	fmt.Println("Decrementing the request count.")
// 	err = t.dbclient.Decrement(t.domain, userID+"_counter")
// 	if err != nil {
// 		// Handle error (e.g., failed to decrement value)
// 		fmt.Printf("Error decrementing counter: %v\n", err)
// 		return false
// 	}

// 	// Allow the request
// 	fmt.Println("Request allowed.")
// 	return true
// }