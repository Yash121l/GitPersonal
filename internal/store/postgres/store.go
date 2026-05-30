package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/yashlunawat/forge/internal/store"
)

const (
	createUserQuery = `
INSERT INTO users (username, password_hash, role)
VALUES ($1, $2, $3)
RETURNING id, username, password_hash, role, created_at`

	getUserByUsernameQuery = `
SELECT id, username, password_hash, role, created_at
FROM users
WHERE lower(username) = lower($1)`

	getUserByIDQuery = `
SELECT id, username, password_hash, role, created_at
FROM users
WHERE id = $1`

	createRepositoryQuery = `
INSERT INTO repositories (
	owner_user_id,
	name,
	description,
	visibility,
	default_branch,
	repo_path
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, description, visibility, default_branch, is_archived, repo_path, size_bytes, last_indexed_at, last_maintained_at, created_at, updated_at`

	getRepositoryByOwnerAndNameQuery = `
SELECT
	r.id,
	u.username,
	r.name,
	r.description,
	r.visibility,
	r.default_branch,
	r.is_archived,
	r.repo_path,
	r.size_bytes,
	r.last_indexed_at,
	r.last_maintained_at,
	r.created_at,
	r.updated_at
FROM repositories r
JOIN users u ON u.id = r.owner_user_id
WHERE lower(u.username) = lower($1)
  AND lower(r.name) = lower($2)
LIMIT 1`

	listRepositoriesByOwnerQuery = `
SELECT
	r.id,
	u.username,
	r.name,
	r.description,
	r.visibility,
	r.default_branch,
	r.is_archived,
	r.repo_path,
	r.size_bytes,
	r.last_indexed_at,
	r.last_maintained_at,
	r.created_at,
	r.updated_at
FROM repositories r
JOIN users u ON u.id = r.owner_user_id
WHERE lower(u.username) = lower($1)
ORDER BY r.name ASC`

	listRepositoriesQuery = `
SELECT
	r.id,
	u.username,
	r.name,
	r.description,
	r.visibility,
	r.default_branch,
	r.is_archived,
	r.repo_path,
	r.size_bytes,
	r.last_indexed_at,
	r.last_maintained_at,
	r.created_at,
	r.updated_at
FROM repositories r
JOIN users u ON u.id = r.owner_user_id
ORDER BY u.username ASC, r.name ASC`

	updateRepositoryStatsQuery = `
UPDATE repositories r
SET
	size_bytes = $3,
	last_indexed_at = COALESCE($4, last_indexed_at),
	last_maintained_at = COALESCE($5, last_maintained_at),
	updated_at = NOW()
FROM users u
WHERE r.owner_user_id = u.id
  AND lower(u.username) = lower($1)
  AND lower(r.name) = lower($2)`

	deleteRepositoryQuery = `
DELETE FROM repositories r
USING users u
WHERE r.owner_user_id = u.id
  AND lower(u.username) = lower($1)
  AND lower(r.name) = lower($2)`

	createSessionQuery = `
INSERT INTO sessions (user_id, token_id, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token_id::text, expires_at, created_at, revoked_at`

	getSessionByTokenIDQuery = `
SELECT id, user_id, token_id::text, expires_at, created_at, revoked_at
FROM sessions
WHERE token_id = $1::uuid`

	revokeSessionQuery = `
UPDATE sessions
SET revoked_at = $2
WHERE token_id = $1::uuid`

	createSSHKeyQuery = `
INSERT INTO ssh_keys (user_id, name, public_key, fingerprint_sha256)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, name, public_key, fingerprint_sha256, created_at, last_used_at`

	getUserBySSHFingerprintQuery = `
SELECT u.id, u.username, u.password_hash, u.role, u.created_at
FROM ssh_keys k
JOIN users u ON u.id = k.user_id
WHERE k.fingerprint_sha256 = $1`

	touchSSHKeyUsageQuery = `
UPDATE ssh_keys
SET last_used_at = $2
WHERE fingerprint_sha256 = $1`
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash, role string) (store.User, error) {
	var user store.User
	err := s.db.QueryRowContext(ctx, createUserQuery, username, passwordHash, role).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return store.User{}, store.ErrAlreadyExists
		}
		return store.User{}, err
	}

	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (store.User, error) {
	var user store.User
	err := s.db.QueryRowContext(ctx, getUserByIDQuery, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.User{}, store.ErrNotFound
		}
		return store.User{}, err
	}

	return user, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (store.User, error) {
	var user store.User
	err := s.db.QueryRowContext(ctx, getUserByUsernameQuery, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.User{}, store.ErrNotFound
		}
		return store.User{}, err
	}

	return user, nil
}

func (s *Store) CreateRepository(ctx context.Context, params store.CreateRepositoryParams) (store.Repository, error) {
	owner, err := s.GetUserByUsername(ctx, params.Owner)
	if err != nil {
		return store.Repository{}, err
	}

	var repository store.Repository
	err = s.db.QueryRowContext(
		ctx,
		createRepositoryQuery,
		owner.ID,
		params.Name,
		params.Description,
		params.Visibility,
		params.DefaultBranch,
		params.RepoPath,
	).Scan(
		&repository.ID,
		&repository.Name,
		&repository.Description,
		&repository.Visibility,
		&repository.DefaultBranch,
		&repository.Archived,
		&repository.RepoPath,
		&repository.SizeBytes,
		&repository.LastIndexedAt,
		&repository.LastMaintainedAt,
		&repository.CreatedAt,
		&repository.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return store.Repository{}, store.ErrAlreadyExists
		}
		return store.Repository{}, err
	}

	repository.Owner = owner.Username
	return repository, nil
}

func (s *Store) GetRepositoryByOwnerAndName(ctx context.Context, owner, name string) (store.Repository, error) {
	var repository store.Repository
	err := s.db.QueryRowContext(ctx, getRepositoryByOwnerAndNameQuery, owner, name).Scan(
		&repository.ID,
		&repository.Owner,
		&repository.Name,
		&repository.Description,
		&repository.Visibility,
		&repository.DefaultBranch,
		&repository.Archived,
		&repository.RepoPath,
		&repository.SizeBytes,
		&repository.LastIndexedAt,
		&repository.LastMaintainedAt,
		&repository.CreatedAt,
		&repository.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.Repository{}, store.ErrNotFound
		}
		return store.Repository{}, err
	}

	return repository, nil
}

func (s *Store) ListRepositories(ctx context.Context) ([]store.Repository, error) {
	rows, err := s.db.QueryContext(ctx, listRepositoriesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRepositories(rows)
}

func (s *Store) ListRepositoriesByOwner(ctx context.Context, owner string) ([]store.Repository, error) {
	rows, err := s.db.QueryContext(ctx, listRepositoriesByOwnerQuery, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRepositories(rows)
}

func (s *Store) UpdateRepositoryStats(ctx context.Context, owner, name string, sizeBytes int64, indexedAt, maintainedAt *time.Time) error {
	result, err := s.db.ExecContext(ctx, updateRepositoryStatsQuery, owner, name, sizeBytes, indexedAt, maintainedAt)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *Store) DeleteRepository(ctx context.Context, owner, name string) error {
	result, err := s.db.ExecContext(ctx, deleteRepositoryQuery, owner, name)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}

func (s *Store) CreateSession(ctx context.Context, params store.CreateSessionParams) (store.Session, error) {
	var session store.Session
	err := s.db.QueryRowContext(ctx, createSessionQuery, params.UserID, params.TokenID, params.ExpiresAt).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenID,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.RevokedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return store.Session{}, store.ErrAlreadyExists
		}
		return store.Session{}, err
	}
	return session, nil
}

func (s *Store) GetSessionByTokenID(ctx context.Context, tokenID string) (store.Session, error) {
	var session store.Session
	err := s.db.QueryRowContext(ctx, getSessionByTokenIDQuery, tokenID).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenID,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.Session{}, store.ErrNotFound
		}
		return store.Session{}, err
	}
	return session, nil
}

func (s *Store) RevokeSession(ctx context.Context, tokenID string, revokedAt time.Time) error {
	result, err := s.db.ExecContext(ctx, revokeSessionQuery, tokenID, revokedAt)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *Store) CreateSSHKey(ctx context.Context, params store.CreateSSHKeyParams) (store.SSHKey, error) {
	var key store.SSHKey
	err := s.db.QueryRowContext(ctx, createSSHKeyQuery, params.UserID, params.Name, params.PublicKey, params.FingerprintSHA256).Scan(
		&key.ID,
		&key.UserID,
		&key.Name,
		&key.PublicKey,
		&key.FingerprintSHA256,
		&key.CreatedAt,
		&key.LastUsedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return store.SSHKey{}, store.ErrAlreadyExists
		}
		return store.SSHKey{}, err
	}
	return key, nil
}

func (s *Store) GetUserBySSHFingerprint(ctx context.Context, fingerprint string) (store.User, error) {
	var user store.User
	err := s.db.QueryRowContext(ctx, getUserBySSHFingerprintQuery, fingerprint).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.User{}, store.ErrNotFound
		}
		return store.User{}, err
	}
	return user, nil
}

func (s *Store) TouchSSHKeyUsage(ctx context.Context, fingerprint string, usedAt time.Time) error {
	result, err := s.db.ExecContext(ctx, touchSSHKeyUsageQuery, fingerprint, usedAt)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *Store) WithRepositoryLease(ctx context.Context, owner, name string, fn func(context.Context) error) error {
	return fn(ctx)
}

func (s *Store) Check(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func scanRepositories(rows *sql.Rows) ([]store.Repository, error) {
	var repositories []store.Repository
	for rows.Next() {
		var repository store.Repository
		if err := rows.Scan(
			&repository.ID,
			&repository.Owner,
			&repository.Name,
			&repository.Description,
			&repository.Visibility,
			&repository.DefaultBranch,
			&repository.Archived,
			&repository.RepoPath,
			&repository.SizeBytes,
			&repository.LastIndexedAt,
			&repository.LastMaintainedAt,
			&repository.CreatedAt,
			&repository.UpdatedAt,
		); err != nil {
			return nil, err
		}
		repositories = append(repositories, repository)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return repositories, nil
}
