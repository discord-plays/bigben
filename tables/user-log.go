package tables

import (
	"github.com/disgoorg/snowflake/v2"
)

type UserLog struct {
	Id  snowflake.ID `xorm:"pk 'id''" csv:"id"`
	Tag string       `xorm:"tag" csv:"tag"`
}
