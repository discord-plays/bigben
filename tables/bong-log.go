package tables

import "time"

type BongLog struct {
	GuildId          string    `xorm:"guild_id"`
	UserId           string    `xorm:"user_id"`
	Timestamp        time.Time `xorm:"timestamp"`
	MessageTimestamp time.Time `xorm:"message_timestamp"`
}
