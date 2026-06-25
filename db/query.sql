-- name: CreateProxy :one
INSERT INTO proxies (address, username, password) 
VALUES ($1, $2, $3) 
ON CONFLICT (address) DO NOTHING 
RETURNING *;

-- name: GetPendingTask :one
SELECT 
    t.*, 
    p.address AS proxy_address, 
    p.username AS proxy_username, 
    p.password AS proxy_password
FROM tasks t
LEFT JOIN proxies p ON t.proxy_id = p.id
WHERE t.status = 'pending' 
ORDER BY t.created_at ASC 
LIMIT 1
FOR UPDATE OF t SKIP LOCKED;

-- name: UpdateTaskStatus :one
UPDATE tasks SET status = $1, updated_at = NOW(), error_message = $2 WHERE id = $3 RETURNING *;