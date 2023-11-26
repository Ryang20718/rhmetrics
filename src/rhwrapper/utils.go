package rhwrapper

import (
	"time"
)

func OneYearApart(dateStr1 string, dateStr2 string) bool {
	date1, _ := time.Parse("2006-01-02", dateStr1)
	date2, _ := time.Parse("2006-01-02", dateStr2)

	// Calculate the duration between the two dates
	duration := date2.Sub(date1)

	// Check if the duration is exactly 365 days
	if duration.Hours() == 365*24 {
		return true
	}
	return false
}
