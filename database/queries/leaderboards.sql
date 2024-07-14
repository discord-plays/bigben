-- name: TotalClicksLeaderboard :many
SELECT user_id, count(user_id) as bong_count
FROM bongs
WHERE guild_id = ?
  and message_id > ?
  and won = true
GROUP BY user_id
ORDER BY bong_count DESC, user_id DESC
LIMIT 10;

-- name: AverageSpeedLeaderboard :many
SELECT user_id, cast(avg(speed) as float) as average_speed
FROM bongs
WHERE guild_id = ?
  and message_id > ?
  and won = true
GROUP BY user_id
ORDER BY average_speed, user_id DESC
LIMIT 10;

-- name: SlowestSpeedLeaderboard :many
SELECT user_id, cast(max(speed) as float) as max_speed
FROM bongs
WHERE guild_id = ?
  and message_id > ?
  and won = true
GROUP BY user_id
ORDER BY max_speed DESC, user_id DESC
LIMIT 10;

-- name: FastestSpeedLeaderboard :many
SELECT user_id, cast(min(speed) as float) as min_speed
FROM bongs
WHERE guild_id = ?
  and message_id > ?
  and won = true
GROUP BY user_id
ORDER BY min_speed, user_id DESC
LIMIT 10;
