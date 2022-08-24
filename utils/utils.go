package utils

import "github.com/bwmarrin/discordgo"

type MainBotInterface interface {
	AppId() string
	GuildId() string
	Session() *discordgo.Session
}
