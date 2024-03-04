package inter

import (
	"github.com/discord-plays/bigben/database"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
)

type MainBotInterface interface {
	AppId() snowflake.ID
	GuildId() snowflake.ID
	Session() bot.Client
	Engine() *database.Queries
}
