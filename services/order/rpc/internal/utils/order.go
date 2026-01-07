package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateOrderNo generates a unique order number
// Format: LG + YYYYMMDDHHmmss + 4-digit random number
// Example: LG20251229162345001
func GenerateOrderNo() string {
	// Use current timestamp
	now := time.Now()
	timestamp := now.Format("20060102150405")

	// Generate 4-digit random number
	random := rand.Intn(10000)

	return fmt.Sprintf("LG%s%04d", timestamp, random)
}

// init initializes random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}
