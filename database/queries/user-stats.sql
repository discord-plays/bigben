-- name: UserStats :one
SELECT count(speed) as total_bongs, cast(avg(speed) as float) as average_speed
FROM bongs
WHERE guild_id = ?
  and user_id = ?
  and won = true
LIMIT 1;
