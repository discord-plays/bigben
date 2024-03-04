// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package database

import (
	snowflake "github.com/disgoorg/snowflake/v2"
)

type Bong struct {
	ID            int32        `json:"id"`
	GuildID       snowflake.ID `json:"guild_id"`
	UserID        snowflake.ID `json:"user_id"`
	MessageID     snowflake.ID `json:"message_id"`
	InteractionID snowflake.ID `json:"interaction_id"`
	Won           bool         `json:"won"`
	Speed         int64        `json:"speed"`
}

type Guild struct {
	ID               snowflake.ID `json:"id"`
	BongChannelID    snowflake.ID `json:"bong_channel_id"`
	BongWebhookID    snowflake.ID `json:"bong_webhook_id"`
	BongWebhookToken string       `json:"bong_webhook_token"`
	BongRoleID       snowflake.ID `json:"bong_role_id"`
	BongEmoji        string       `json:"bong_emoji"`
}

type Role struct {
	ID        int32        `json:"id"`
	GuildID   snowflake.ID `json:"guild_id"`
	MessageID snowflake.ID `json:"message_id"`
	RoleID    snowflake.ID `json:"role_id"`
	UserID    snowflake.ID `json:"user_id"`
}

type User struct {
	ID  snowflake.ID `json:"id"`
	Tag string       `json:"tag"`
}
