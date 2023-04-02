package utils

import (
	"fmt"
	time "time"
)

const defaultBongText = "Bong"

type BongOption struct {
	T time.Time
	S string
	A bool
}

type DynamicBongOption func(now time.Time) BongOption

func GetBongOptions() []BongOption {
	now := time.Now().UTC()
	bongs := make([]BongOption, len(staticBongs))
	for _, i := range staticBongs {
		bongs = append(bongs, BongOption{T: i.T.AddDate(now.Year(), 0, 0), S: i.S})
	}
	for _, i := range dynamicBong {
		bongs = append(bongs, i(now))
	}
	return bongs
}

func GetBongTitle(t time.Time) BongOption {
	options := GetBongOptions()
	for _, i := range options {
		if EqualDate(t, i.T) {
			return i
		}
	}
	return BongOption{T: t, S: defaultBongText}
}

var staticBongs = []BongOption{
	{T: time.Date(0, time.February, 14, 0, 0, 0, 0, time.UTC), S: "Valentine's Bong 💝"},
	{T: time.Date(0, time.April, 22, 0, 0, 0, 0, time.UTC), S: "Earth Bong 🌍"},
	{T: time.Date(0, time.July, 2, 0, 0, 0, 0, time.UTC), S: "Midway Bong"},
	{T: time.Date(0, time.August, 28, 0, 0, 0, 0, time.UTC), S: "Melon Bong 🍉"},
	{T: time.Date(0, time.October, 31, 0, 0, 0, 0, time.UTC), S: "Spooky Bong 🎃"},
	{T: time.Date(0, time.December, 25, 0, 0, 0, 0, time.UTC), S: "Christmas Bong 🎄"},
}

var dynamicBong = []DynamicBongOption{
	func(now time.Time) BongOption {
		return BongOption{T: time.Date(now.Year(), time.April, 1, 0, 0, 0, 0, time.UTC), S: "Bing", A: true}
	},
	func(now time.Time) BongOption {
		return BongOption{T: time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.UTC), S: fmt.Sprintf("%d Bong 🎆", now.Year())}
	},
	func(now time.Time) BongOption {
		// https://rosettacode.org/wiki/Holidays_related_to_Easter#Go
		y := now.Year()
		c := y / 100
		n := mod(y, 19)
		i := mod(c-c/4-(c-(c-17)/25)/3+19*n+15, 30)
		i -= (i / 28) * (1 - (i/28)*(29/(i+1))*((21-n)/11))
		l := i - mod(y+y/4+i+2-c+c/4, 7)
		m := 3 + (l+40)/44
		d := l + 28 - 31*(m/4)
		return BongOption{T: time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC), S: "Easter Bong 🐰"}
	},
}
