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
		a := now.Year() % 19
		b := now.Year() / 100
		c := now.Year() % 100
		d := b / 4
		e := b % 4
		g := (8*b + 13) / 15
		h := (19*a + b - d - g + 15) % 30
		j := c / 4
		k := c % 4
		m := (a + 11*h) / 319
		r := (2*e + 2*j - k - h + m + 32) % 7
		month := (h - m + r + 90) / 25
		day := (h - m + r + month + 19) % 32
		return BongOption{T: time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, time.UTC), S: "Easter Bong 🐰"}
	},
}
