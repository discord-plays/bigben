package tables

import "github.com/MrMelon54/BigBen/utils"

type RoleLog struct {
	Id        int64               `xorm:"pk autoincr"`
	GuildId   utils.XormSnowflake `xorm:"guild_id"`
	MessageId utils.XormSnowflake `xorm:"message_id"`
	RoleId    utils.XormSnowflake `xorm:"role_id"`
	UserId    utils.XormSnowflake `xorm:"user_id"`
}
