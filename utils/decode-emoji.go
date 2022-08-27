package utils

import (
	"github.com/Succo/emoji"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"regexp"
)

var decodeDiscordEmojiSource = regexp.MustCompile("^<(a?):(.+?):(\\d{18,19})>")

func DecodeDiscordEmoji(a string) (string, bool, int) {
	decode := decodeDiscordEmojiSource.FindString(a)
	if decode == "" {
		return emoji.DecodeString(a)
	}
	return decode, true, len(decode)
}

func DecodeAllDiscordEmoji(a string) (emojiStr []string) {
	for {
		g, ok, n := DecodeDiscordEmoji(a)
		if n == 0 {
			break
		}
		if ok {
			emojiStr = append(emojiStr, g)
		}
		a = a[n:]
	}
	return
}

func ConvertToComponentEmoji(a string) discordgo.ComponentEmoji {
	sub := decodeDiscordEmojiSource.FindStringSubmatch(a)
	if len(sub) == 4 {
		return discordgo.ComponentEmoji{
			Animated: sub[1] == "a",
			Name:     sub[2],
			ID:       sub[3],
		}
	}
	decode, ok, _ := emoji.DecodeString(a)
	if ok {
		return discordgo.ComponentEmoji{Name: decode}
	}
	return discordgo.ComponentEmoji{}
}

func RandomEmoji(a string) string {
	emojis := DecodeAllDiscordEmoji(a)
	n := rand.Intn(len(emojis))
	return emojis[n]
}
