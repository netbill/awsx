-- name: CreateUploadSession :one
INSERT INTO upload_sessions (
    id,
    owner_id,
    purpose,
    expires_at,
    max_files
) VALUES (
    sqlc.arg(id)::uuid,
    sqlc.arg(owner_id)::uuid,
    sqlc.arg(purpose)::text,
    sqlc.arg(expires_at)::timestamptz,
    sqlc.arg(max_files)::int
)
RETURNING id, owner_id, purpose, created_at, expires_at, max_files;

-- name: GetUploadSessionByID :one
SELECT id, owner_id, purpose, created_at, expires_at, max_files
FROM upload_sessions
WHERE id = sqlc.arg(id)::uuid;

-- name: GetActiveUploadSessionByOwnerPurpose :one
SELECT id, owner_id, purpose, created_at, expires_at, max_files
FROM upload_sessions
WHERE owner_id = sqlc.arg(owner_id)::uuid
AND purpose  = sqlc.arg(purpose)::text
ORDER BY created_at DESC
LIMIT 1;


-- name: CreateUploadFile :one
INSERT INTO upload_files (
    id,
    session_id,
    object_key
) VALUES (
    sqlc.arg(id)::uuid,
    sqlc.arg(session_id)::uuid,
    sqlc.arg(object_key)::text
)
RETURNING id, session_id, object_key, created_at;

-- name: ListUploadFilesBySession :many
SELECT id, session_id, object_key, created_at
FROM upload_files
WHERE session_id = sqlc.arg(session_id)::uuid
ORDER BY created_at ASC;

-- name: GetUploadFileByID :one
SELECT id, session_id, object_key, created_at
FROM upload_files
WHERE id = sqlc.arg(id)::uuid;

-- name: DeleteUploadSession :exec
DELETE FROM upload_sessions
WHERE id = sqlc.arg(id)::uuid;
