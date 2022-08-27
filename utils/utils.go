package utils

import (
	"github.com/MrMelon54/BigBen/tables"
	"github.com/bwmarrin/discordgo"
	"time"
	"xorm.io/xorm"
)

type MainBotInterface interface {
	AppId() string
	GuildId() string
	Session() *discordgo.Session
	GetGuildSettings(guildId string) (tables.GuildSettings, error)
	PutGuildSettings(guildSettings tables.GuildSettings) error
	Engine() *xorm.Engine
}

func GetStartOfHourTime() time.Time {
	n := time.Now().UTC()
	return time.Date(n.Year(), n.Month(), n.Day(), n.Hour(), 0, 0, 0, time.UTC)
}

func EqualDate(a, b time.Time) bool {
	a1, a2, a3 := a.Date()
	b1, b2, b3 := b.Date()
	return a1 == b1 && a2 == b2 && a3 == b3
}
