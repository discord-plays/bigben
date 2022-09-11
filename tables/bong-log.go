package tables

import "time"

type BongLog struct {
	Id            int64     `xorm:"pk autoincr"`
	GuildId       string    `xorm:"guild_id"`
	UserId        string    `xorm:"user_id"`
	Timestamp     time.Time `xorm:"timestamp"`
	TimeDetail    int64     `xorm:"time_detail"`
	MsgTimestamp  time.Time `xorm:"message_timestamp"`
	MsgTimeDetail int64     `xorm:"message_time_detail"`
}
