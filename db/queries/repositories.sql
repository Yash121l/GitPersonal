-- name: CreateRepository :one
INSERT INTO repositories (
    owner_user_id,
    name,
    description,
    visibility,
    default_branch,
    repo_path
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING id, owner_user_id, owner_org_id, name, description, visibility, default_branch, is_archived, repo_path, size_bytes, created_at, updated_at;

-- name: ListRepositoriesByOwnerUser :many
SELECT id, owner_user_id, owner_org_id, name, description, visibility, default_branch, is_archived, repo_path, size_bytes, created_at, updated_at
FROM repositories
WHERE owner_user_id = $1
ORDER BY name ASC;

-- name: DeleteRepositoryByOwnerUser :exec
DELETE FROM repositories
WHERE owner_user_id = $1
  AND lower(name) = lower($2);

