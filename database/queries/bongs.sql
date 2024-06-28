-- name: AddBong :exec
INSERT INTO bongs (guild_id, user_id, message_id, interaction_id, won, speed)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ReplaceUser :exec
REPLACE INTO users (id, tag)
VALUES (?, ?);
