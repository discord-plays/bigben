-- name: GetRoleLogs :many
SELECT *
FROM roles
WHERE guild_id = ?
  AND message_id != ?;

-- name: DeleteRoles :exec
DELETE
FROM roles
WHERE id IN (sqlc.slice(role_ids));

-- name: AddRole :exec
INSERT INTO roles (guild_id, message_id, role_id, user_id)
VALUES (?, ?, ?, ?);
