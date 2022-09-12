package utils

import (
	"time"
)

func GetStartOfHourTime() time.Time {
	n := time.Now().UTC()
	return time.Date(n.Year(), n.Month(), n.Day(), n.Hour(), 0, 0, 0, time.UTC)
}

func EqualDate(a, b time.Time) bool {
	a1, a2, a3 := a.Date()
	b1, b2, b3 := b.Date()
	return a1 == b1 && a2 == b2 && a3 == b3
}
