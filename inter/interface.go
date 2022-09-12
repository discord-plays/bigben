package inter

import (
	"github.com/MrMelon54/BigBen/tables"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
	"xorm.io/xorm"
)

type MainBotInterface interface {
	AppId() snowflake.ID
	GuildId() snowflake.ID
	Session() bot.Client
	GetGuildSettings(guildId snowflake.ID) (tables.GuildSettings, error)
	PutGuildSettings(guildSettings tables.GuildSettings) error
	Engine() *xorm.Engine
}
