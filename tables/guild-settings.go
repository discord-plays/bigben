package tables

type GuildSettings struct {
	GuildId          string `xorm:"guild_id"`
	BongChannelId    string `xorm:"bong_channel_id"`
	BongWebhookId    string `xorm:"bong_webhook_id"`
	BongWebhookToken string `xorm:"bong_webhook_token"`
	BongRoleId       string `xorm:"bong_role_id"`
	BongEmoji        string `xorm:"bong_emoji"`
}
