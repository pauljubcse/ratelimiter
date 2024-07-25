package ratelimiter

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pauljubcse/kvsclient"
)

type SlidingWindowLogsLimiter struct {
	dbclient          *kvsclient.Client
	domain            string
	intervalInSeconds int64
	maximumRequests   int64
}

// RateLimitUsingSlidingLogs implements the sliding window log rate limiter
func (s *SlidingWindowLogsLimiter) Allow(userID string) bool {
	currentTime := time.Now().UnixNano()
	lastWindowTime := currentTime - s.intervalInSeconds*1000000000

	slkey := userID + ":sl"
	// Count the number of logs in the current window
	r1, err := s.dbclient.RankInSkipList(s.domain, slkey, strconv.FormatInt(lastWindowTime, 10))
	if err != nil {
		fmt.Printf("Failed to get rank: %v\n", err)
		//fmt.Printf("Failed to get rank: %v\n", err)
		s.dbclient.InsertToSkipList(s.domain, slkey, "0", "") //Will initialise the skiplist
		//return false
		r1="1"
		//return false
	}
	rankLastWindow, err := strconv.ParseInt(r1, 10, 64)
	if err != nil {
		fmt.Printf("Failed to parse rank: %v\n", err)
		return false
	}

	r2, err := s.dbclient.RankInSkipList(s.domain, slkey, strconv.FormatInt(currentTime, 10))
	if err != nil {
		fmt.Printf("Failed to get rank: %v\n", err)
		//s.dbclient.InsertToSkipList(s.domain)
		return false
	}
	rankCurrentTime, err := strconv.ParseInt(r2, 10, 64)
	if err != nil {
		fmt.Printf("Failed to parse rank: %v\n", err)
		return false
	}
	fmt.Printf("Last: %d\tCurrent: %d\n", lastWindowTime, currentTime)
	fmt.Printf("Rank Last: %d\tRank Current: %d\n", rankLastWindow, rankCurrentTime)
	// Check if the count exceeds the maximum allowed requests
	if rankCurrentTime-rankLastWindow >= s.maximumRequests {
		return false
	}

	// Add the current timestamp to the skip list
	err = s.dbclient.InsertToSkipList(s.domain, slkey, strconv.FormatInt(currentTime, 10), "")
	if err != nil {
		fmt.Printf("Failed to add timestamp: %v\n", err)
		return false
	}
	l, _ := s.dbclient.RankInSkipList(s.domain, slkey, strconv.FormatInt(currentTime+1, 10))
	fmt.Printf("Length: %s\n", l)

	// Clean up old timestamps
	err = s.dbclient.DeleteRangeFromSkipList(s.domain, slkey, strconv.FormatInt(1, 10), strconv.FormatInt(lastWindowTime, 10))
	if err != nil {
		fmt.Printf("Failed to delete old timestamps: %v\n", err)
	}
	l, _ = s.dbclient.RankInSkipList(s.domain, slkey, strconv.FormatInt(currentTime+1, 10))
	fmt.Printf("Length after deletion: %s\n", l)

	return true
}