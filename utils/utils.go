package utils

import (
	"github.com/MrMelon54/BigBen/tables"
	"github.com/Succo/emoji"
	"github.com/bwmarrin/discordgo"
	"regexp"
)

type MainBotInterface interface {
	AppId() string
	GuildId() string
	Session() *discordgo.Session
	GetGuildSettings(guildId string) (tables.GuildSettings, error)
	PutGuildSettings(guildSettings tables.GuildSettings) error
}

var decodeDiscordEmojiSource = regexp.MustCompile("/^<a?:.+?:\\d{18}>")

func DecodeDiscordEmoji(a string) (string, bool, int) {
	decode := decodeDiscordEmojiSource.FindString(a)
	if decode == "" {
		return emoji.DecodeString(a)
	}
	return decode, true, len(decode)
}

func DecodeAllDiscordEmoji(a string) (emojiStr []string) {
	for {
		decode, ok, n := DecodeDiscordEmoji(a)
		if !ok {
			break
		}
		emojiStr = append(emojiStr, decode)
		a = a[n:]
	}
	return
}
