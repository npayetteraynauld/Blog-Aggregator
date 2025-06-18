-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
  INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
  VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
  )
  RETURNING *
)
SELECT 
  inserted_feed_follow.*,
  feeds.name as feed_name,
  users.name as user_name
FROM inserted_feed_follow
INNER JOIN feeds
  ON inserted_feed_follow.feed_id = feeds.id
INNER JOIN users
  ON inserted_feed_follow.user_id = users.id;

-- name: GetFeedFollowsForUser :many
SELECT 
  feed_follows.*,
  feeds.name as feed_name,
  users.name as user_name
From feed_follows
INNER JOIN feeds
  ON feed_follows.feed_id = feeds.id
INNER JOIN users
  ON feed_follows.user_id = users.id
WHERE $1 = feed_follows.user_id;

-- name: Unfollow :exec
DELETE from feed_follows
WHERE $1 = feed_follows.user_id
AND $2 = feed_follows.feed_id;
