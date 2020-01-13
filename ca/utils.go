package ca

import "time"

func timeNotAfter() time.Time {
	now := time.Now()
	return time.Date(now.Year()+10, now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
