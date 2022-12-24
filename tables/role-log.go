package tables

import (
	"github.com/disgoorg/snowflake/v2"
)

type RoleLog struct {
	Id        int64        `xorm:"pk autoincr" csv:"id"`
	GuildId   snowflake.ID `xorm:"guild_id" csv:"guild_id"`
	MessageId snowflake.ID `xorm:"message_id" csv:"message_id"`
	RoleId    snowflake.ID `xorm:"role_id" csv:"role_id"`
	UserId    snowflake.ID `xorm:"user_id" csv:"user_id"`
}
