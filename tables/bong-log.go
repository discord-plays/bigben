package tables

import (
	"github.com/disgoorg/snowflake/v2"
)

type BongLog struct {
	Id      int64        `xorm:"pk autoincr"`
	GuildId snowflake.ID `xorm:"guild_id"`
	UserId  snowflake.ID `xorm:"user_id"`
	MsgId   snowflake.ID `xorm:"msg_id"`
	InterId snowflake.ID `xorm:"inter_id"`
	Won     *bool        `xorm:"won"`
	Speed   int64        `xorm:"speed"`
}
