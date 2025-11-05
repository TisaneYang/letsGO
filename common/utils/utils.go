package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

// ========================================
// Password Utilities
// ========================================

// HashPassword generates MD5 hash of password with salt
// In production, use bcrypt instead of MD5!
func HashPassword(password, salt string) string {
	h := md5.New()
	h.Write([]byte(password + salt))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateSalt generates a random salt string
func GenerateSalt() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ========================================
// Order Number Generation
// ========================================

// GenerateOrderNo generates unique order number
// Format: LG + YYYYMMDDHHMMSS + 6-digit random number
// Example: LG20250104123456123456
func GenerateOrderNo() string {
	now := time.Now()
	dateStr := now.Format("20060102150405")
	rand.Seed(now.UnixNano())
	randomNum := rand.Intn(999999)
	return fmt.Sprintf("LG%s%06d", dateStr, randomNum)
}

// ========================================
// Payment Number Generation
// ========================================

// GeneratePaymentNo generates unique payment number
// Format: PAY + YYYYMMDDHHMMSS + 6-digit random number
func GeneratePaymentNo() string {
	now := time.Now()
	dateStr := now.Format("20060102150405")
	rand.Seed(now.UnixNano())
	randomNum := rand.Intn(999999)
	return fmt.Sprintf("PAY%s%06d", dateStr, randomNum)
}

// ========================================
// Time Utilities
// ========================================

// GetCurrentTimestamp returns current Unix timestamp in seconds
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// GetCurrentMilliTimestamp returns current Unix timestamp in milliseconds
func GetCurrentMilliTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// ========================================
// Pagination Utilities
// ========================================

// CalculateOffset calculates database offset for pagination
func CalculateOffset(page, pageSize int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * pageSize
}

// ValidatePageParams validates and corrects page parameters
func ValidatePageParams(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // Max 100 items per page
	}
	return page, pageSize
}
