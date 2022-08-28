package tables

type RoleLog struct {
	Id        int64  `xorm:"pk autoincr"`
	GuildId   string `xorm:"guild_id"`
	MessageId string `xorm:"message_id"`
	RoleId    string `xorm:"role_id"`
	UserId    string `xorm:"user_id"`
}
