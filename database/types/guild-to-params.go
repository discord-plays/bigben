package types

import "github.com/discord-plays/bigben/database"

func ConvertGuildToParams(guild database.Guild) database.UpdateGuildParams {
	return database.UpdateGuildParams{
		ID:               guild.ID,
		BongChannelID:    guild.BongChannelID,
		BongWebhookID:    guild.BongWebhookID,
		BongWebhookToken: guild.BongWebhookToken,
		BongRoleID:       guild.BongRoleID,
		BongEmoji:        guild.BongEmoji,
	}
}
