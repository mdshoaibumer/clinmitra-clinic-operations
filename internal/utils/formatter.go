package utils

import (
	"fmt"
	"time"
)

// FormatPaise converts paise to formatted rupee string
func FormatPaise(paise int64) string {
	rupees := paise / 100
	remainingPaise := paise % 100
	return fmt.Sprintf("₹%d.%02d", rupees, remainingPaise)
}

// FormatDate formats a date string for display
func FormatDate(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("02 Jan 2006")
}

// TodayDate returns today's date in YYYY-MM-DD format
func TodayDate() string {
	return time.Now().Format("2006-01-02")
}

// CurrentTime returns current time in HH:MM format
func CurrentTime() string {
	return time.Now().Format("15:04")
}

// CurrentMonth returns YYMM for invoice numbering
func CurrentMonth() string {
	return time.Now().Format("0601")
}
