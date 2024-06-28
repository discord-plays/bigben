-- name: GetAllGuilds :many
SELECT *
FROM guilds;

-- name: GetGuild :one
SELECT *
FROM guilds
WHERE id = ?;

-- name: UpdateGuild :exec
REPLACE INTO guilds (id, bong_channel_id, bong_webhook_id, bong_webhook_token, bong_role_id, bong_emoji)
VALUES (?, ?, ?, ?, ?, ?);
