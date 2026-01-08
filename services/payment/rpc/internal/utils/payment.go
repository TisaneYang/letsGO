package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GeneratePaymentNo generates a unique payment number
// Format: PAY + YYYYMMDDHHmmss + 8-digit random number
// Example: PAY20260108123456789012
func GeneratePaymentNo() string {
	// Use current timestamp
	now := time.Now()
	timestamp := now.Format("20060102150405")

	// Generate 8-digit random number
	random := rand.Intn(100000000)

	return fmt.Sprintf("PAY%s%08d", timestamp, random)
}

// init initializes random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}
