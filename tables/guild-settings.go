package tables

import (
	"github.com/disgoorg/snowflake/v2"
)

type GuildSettings struct {
	Id               int64        `xorm:"pk autoincr" csv:"id"`
	GuildId          snowflake.ID `xorm:"guild_id" csv:"guild_id"`
	BongChannelId    snowflake.ID `xorm:"bong_channel_id" csv:"bong_channel_id"`
	BongWebhookId    snowflake.ID `xorm:"bong_webhook_id" csv:"bong_webhook_id"`
	BongWebhookToken string       `xorm:"bong_webhook_token" csv:"bong_webhook_token"`
	BongRoleId       snowflake.ID `xorm:"bong_role_id" csv:"bong_role_id"`
	BongEmoji        string       `xorm:"bong_emoji" csv:"bong_emoji"`
}
