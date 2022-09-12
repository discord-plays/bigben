package tables

import (
	"github.com/MrMelon54/BigBen/utils"
)

type BongLog struct {
	Id      int64               `xorm:"pk autoincr"`
	GuildId utils.XormSnowflake `xorm:"guild_id"`
	UserId  utils.XormSnowflake `xorm:"user_id"`
	MsgId   utils.XormSnowflake `xorm:"msg_id"`
	InterId utils.XormSnowflake `xorm:"inter_id"`
}
