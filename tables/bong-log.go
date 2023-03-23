package tables

import (
	"github.com/disgoorg/snowflake/v2"
)

type BongLog struct {
	Id      int64        `xorm:"pk autoincr" csv:"id"`
	GuildId snowflake.ID `xorm:"guild_id" csv:"guild_id"`
	UserId  snowflake.ID `xorm:"user_id" csv:"user_id"`
	MsgId   snowflake.ID `xorm:"msg_id" csv:"msg_id"`
	InterId snowflake.ID `xorm:"inter_id unique" csv:"inter_id"`
	Won     *bool        `xorm:"won" csv:"won"`
	Speed   int64        `xorm:"speed" csv:"speed"`
}
