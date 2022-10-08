package tables

import (
	"github.com/MrMelon54/bigben/utils"
)

type GuildSettings struct {
	Id               int64               `xorm:"pk autoincr"`
	GuildId          utils.XormSnowflake `xorm:"guild_id"`
	BongChannelId    utils.XormSnowflake `xorm:"bong_channel_id"`
	BongWebhookId    utils.XormSnowflake `xorm:"bong_webhook_id"`
	BongWebhookToken string              `xorm:"bong_webhook_token"`
	BongRoleId       utils.XormSnowflake `xorm:"bong_role_id"`
	BongEmoji        string              `xorm:"bong_emoji"`
}
