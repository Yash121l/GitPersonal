package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

	createOrganizationQuery = `
INSERT INTO organizations (slug, display_name, description, created_by)
VALUES ($1, $2, $3, $4)
RETURNING id, slug, display_name, description, created_by, created_at`

	getOrganizationBySlugQuery = `
SELECT id, slug, display_name, description, created_by, created_at
FROM organizations
WHERE lower(slug) = lower($1)`

	listOrganizationsByMemberQuery = `
SELECT
	o.id,
	o.slug,
	o.display_name,
	m.user_id,
	u.username,
	m.role,
	m.created_at
FROM org_members m
JOIN organizations o ON o.id = m.organization_id
JOIN users u ON u.id = m.user_id
WHERE m.user_id = $1
ORDER BY lower(o.slug) ASC`

	addOrganizationMemberQuery = `
INSERT INTO org_members (organization_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING created_at`

	getOrganizationMembershipQuery = `
SELECT
	o.id,
	o.slug,
	o.display_name,
	m.user_id,
	u.username,
	m.role,
	m.created_at
FROM org_members m
JOIN organizations o ON o.id = m.organization_id
JOIN users u ON u.id = m.user_id
WHERE lower(o.slug) = lower($1)
  AND m.user_id = $2`

	createRepositoryQuery = `
INSERT INTO repositories (
	owner_user_id,
	owner_org_id,
	name,
	description,
	visibility,
	default_branch,
	repo_path
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, name, description, visibility, default_branch, is_archived, repo_path, size_bytes, last_indexed_at, last_maintained_at, created_at, updated_at`

	repositorySelectClause = `
SELECT
	r.id,
	COALESCE(u.username, o.slug) AS owner,
	CASE
		WHEN r.owner_org_id IS NOT NULL THEN 'organization'
		ELSE 'user'
	END AS owner_type,
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
LEFT JOIN users u ON u.id = r.owner_user_id
LEFT JOIN organizations o ON o.id = r.owner_org_id`

	getRepositoryByOwnerAndNameQuery = repositorySelectClause + `
WHERE lower(COALESCE(u.username, o.slug)) = lower($1)
  AND lower(r.name) = lower($2)
LIMIT 1`

	listRepositoriesByOwnerQuery = repositorySelectClause + `
WHERE lower(COALESCE(u.username, o.slug)) = lower($1)
ORDER BY lower(COALESCE(u.username, o.slug)) ASC, owner_type ASC, lower(r.name) ASC`

	listRepositoriesQuery = repositorySelectClause + `
ORDER BY lower(COALESCE(u.username, o.slug)) ASC, owner_type ASC, lower(r.name) ASC`

	listRepositoriesForUserQuery = repositorySelectClause + `
LEFT JOIN org_members om
	ON om.organization_id = r.owner_org_id
	AND om.user_id = $1
LEFT JOIN repo_collaborators rc
	ON rc.repository_id = r.id
	AND rc.user_id = $1
WHERE r.owner_user_id = $1
   OR om.user_id IS NOT NULL
   OR rc.user_id IS NOT NULL
GROUP BY
	r.id,
	u.username,
	o.slug,
	r.owner_org_id,
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
ORDER BY lower(COALESCE(u.username, o.slug)) ASC, owner_type ASC, lower(r.name) ASC`

	updateRepositoryStatsByIDQuery = `
UPDATE repositories
SET
	size_bytes = $2,
	last_indexed_at = COALESCE($3, last_indexed_at),
	last_maintained_at = COALESCE($4, last_maintained_at),
	updated_at = NOW()
WHERE id = $1`

	deleteRepositoryByIDQuery = `
DELETE FROM repositories
WHERE id = $1`

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

	createRepositoryCollaboratorQuery = `
INSERT INTO repo_collaborators (repository_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING created_at`

	getRepositoryCollaboratorQuery = `
SELECT rc.repository_id, rc.user_id, u.username, rc.role, rc.created_at
FROM repo_collaborators rc
JOIN users u ON u.id = rc.user_id
WHERE rc.repository_id = $1
  AND rc.user_id = $2`

	organizationSlugExistsQuery = `
SELECT EXISTS (
	SELECT 1
	FROM organizations
	WHERE lower(slug) = lower($1)
)`

	usernameExistsQuery = `
SELECT EXISTS (
	SELECT 1
	FROM users
	WHERE lower(username) = lower($1)
)`

	acquireRepositoryLeaseQuery = `
SELECT pg_advisory_xact_lock($1)`
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash, role string) (store.User, error) {
	collides, err := s.organizationSlugExists(ctx, username)
	if err != nil {
		return store.User{}, err
	}
	if collides {
		return store.User{}, store.ErrAlreadyExists
	}

	var user store.User
	err = s.db.QueryRowContext(ctx, createUserQuery, username, passwordHash, role).Scan(
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

func (s *Store) CreateOrganization(ctx context.Context, params store.CreateOrganizationParams) (store.Organization, error) {
	collides, err := s.usernameExists(ctx, params.Slug)
	if err != nil {
		return store.Organization{}, err
	}
	if collides {
		return store.Organization{}, store.ErrAlreadyExists
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return store.Organization{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var organization store.Organization
	err = tx.QueryRowContext(
		ctx,
		createOrganizationQuery,
		params.Slug,
		params.DisplayName,
		params.Description,
		params.CreatedBy,
	).Scan(
		&organization.ID,
		&organization.Slug,
		&organization.DisplayName,
		&organization.Description,
		&organization.CreatedBy,
		&organization.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return store.Organization{}, store.ErrAlreadyExists
		}
		return store.Organization{}, err
	}

	if err := tx.QueryRowContext(
		ctx,
		addOrganizationMemberQuery,
		organization.ID,
		params.CreatedBy,
		store.OrganizationRoleOwner,
	).Scan(new(time.Time)); err != nil {
		if isUniqueViolation(err) {
			return store.Organization{}, store.ErrAlreadyExists
		}
		return store.Organization{}, err
	}

	if err := tx.Commit(); err != nil {
		return store.Organization{}, err
	}

	return organization, nil
}

func (s *Store) GetOrganizationBySlug(ctx context.Context, slug string) (store.Organization, error) {
	var organization store.Organization
	err := s.db.QueryRowContext(ctx, getOrganizationBySlugQuery, slug).Scan(
		&organization.ID,
		&organization.Slug,
		&organization.DisplayName,
		&organization.Description,
		&organization.CreatedBy,
		&organization.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.Organization{}, store.ErrNotFound
		}
		return store.Organization{}, err
	}
	return organization, nil
}

func (s *Store) ListOrganizationsByMember(ctx context.Context, userID int64) ([]store.OrganizationMembership, error) {
	rows, err := s.db.QueryContext(ctx, listOrganizationsByMemberQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []store.OrganizationMembership
	for rows.Next() {
		var membership store.OrganizationMembership
		if err := rows.Scan(
			&membership.OrganizationID,
			&membership.OrganizationSlug,
			&membership.OrganizationDisplayName,
			&membership.UserID,
			&membership.Username,
			&membership.Role,
			&membership.CreatedAt,
		); err != nil {
			return nil, err
		}
		memberships = append(memberships, membership)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return memberships, nil
}

func (s *Store) AddOrganizationMember(ctx context.Context, params store.AddOrganizationMemberParams) (store.OrganizationMembership, error) {
	role, err := normalizeOrganizationRole(params.Role)
	if err != nil {
		return store.OrganizationMembership{}, err
	}

	organization, err := s.GetOrganizationBySlug(ctx, params.OrganizationSlug)
	if err != nil {
		return store.OrganizationMembership{}, err
	}

	var user store.User
	user, err = s.GetUserByUsername(ctx, params.Username)
	if err != nil {
		return store.OrganizationMembership{}, err
	}

	var createdAt time.Time
	if err := s.db.QueryRowContext(
		ctx,
		addOrganizationMemberQuery,
		organization.ID,
		user.ID,
		role,
	).Scan(&createdAt); err != nil {
		if isUniqueViolation(err) {
			return store.OrganizationMembership{}, store.ErrAlreadyExists
		}
		return store.OrganizationMembership{}, err
	}

	return store.OrganizationMembership{
		OrganizationID:          organization.ID,
		OrganizationSlug:        organization.Slug,
		OrganizationDisplayName: organization.DisplayName,
		UserID:                  user.ID,
		Username:                user.Username,
		Role:                    role,
		CreatedAt:               createdAt,
	}, nil
}

func (s *Store) GetOrganizationMembership(ctx context.Context, organizationSlug string, userID int64) (store.OrganizationMembership, error) {
	var membership store.OrganizationMembership
	err := s.db.QueryRowContext(ctx, getOrganizationMembershipQuery, organizationSlug, userID).Scan(
		&membership.OrganizationID,
		&membership.OrganizationSlug,
		&membership.OrganizationDisplayName,
		&membership.UserID,
		&membership.Username,
		&membership.Role,
		&membership.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.OrganizationMembership{}, store.ErrNotFound
		}
		return store.OrganizationMembership{}, err
	}
	return membership, nil
}

func (s *Store) CreateRepository(ctx context.Context, params store.CreateRepositoryParams) (store.Repository, error) {
	ownerType, err := normalizeOwnerType(params.OwnerType)
	if err != nil {
		return store.Repository{}, err
	}

	var ownerUserID sql.NullInt64
	var ownerOrgID sql.NullInt64

	switch ownerType {
	case store.OwnerTypeUser:
		owner, err := s.GetUserByUsername(ctx, params.Owner)
		if err != nil {
			return store.Repository{}, err
		}
		ownerUserID = sql.NullInt64{Int64: owner.ID, Valid: true}
	case store.OwnerTypeOrganization:
		organization, err := s.GetOrganizationBySlug(ctx, params.Owner)
		if err != nil {
			return store.Repository{}, err
		}
		ownerOrgID = sql.NullInt64{Int64: organization.ID, Valid: true}
	}

	var repository store.Repository
	err = s.db.QueryRowContext(
		ctx,
		createRepositoryQuery,
		ownerUserID,
		ownerOrgID,
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

	repository.Owner = params.Owner
	repository.OwnerType = ownerType
	return repository, nil
}

func (s *Store) GetRepositoryByOwnerAndName(ctx context.Context, owner, name string) (store.Repository, error) {
	var repository store.Repository
	err := s.db.QueryRowContext(ctx, getRepositoryByOwnerAndNameQuery, owner, name).Scan(
		&repository.ID,
		&repository.Owner,
		&repository.OwnerType,
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

func (s *Store) ListRepositoriesForUser(ctx context.Context, userID int64) ([]store.Repository, error) {
	rows, err := s.db.QueryContext(ctx, listRepositoriesForUserQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRepositories(rows)
}

func (s *Store) AddRepositoryCollaborator(ctx context.Context, params store.AddRepositoryCollaboratorParams) (store.RepositoryCollaborator, error) {
	role, err := normalizeRepositoryRole(params.Role)
	if err != nil {
		return store.RepositoryCollaborator{}, err
	}

	repository, err := s.GetRepositoryByOwnerAndName(ctx, params.Owner, params.RepoName)
	if err != nil {
		return store.RepositoryCollaborator{}, err
	}

	user, err := s.GetUserByUsername(ctx, params.Username)
	if err != nil {
		return store.RepositoryCollaborator{}, err
	}

	var createdAt time.Time
	if err := s.db.QueryRowContext(
		ctx,
		createRepositoryCollaboratorQuery,
		repository.ID,
		user.ID,
		role,
	).Scan(&createdAt); err != nil {
		if isUniqueViolation(err) {
			return store.RepositoryCollaborator{}, store.ErrAlreadyExists
		}
		return store.RepositoryCollaborator{}, err
	}

	return store.RepositoryCollaborator{
		RepositoryID: repository.ID,
		UserID:       user.ID,
		Username:     user.Username,
		Role:         role,
		CreatedAt:    createdAt,
	}, nil
}

func (s *Store) GetRepositoryCollaborator(ctx context.Context, owner, repoName string, userID int64) (store.RepositoryCollaborator, error) {
	repository, err := s.GetRepositoryByOwnerAndName(ctx, owner, repoName)
	if err != nil {
		return store.RepositoryCollaborator{}, err
	}

	var collaborator store.RepositoryCollaborator
	err = s.db.QueryRowContext(ctx, getRepositoryCollaboratorQuery, repository.ID, userID).Scan(
		&collaborator.RepositoryID,
		&collaborator.UserID,
		&collaborator.Username,
		&collaborator.Role,
		&collaborator.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.RepositoryCollaborator{}, store.ErrNotFound
		}
		return store.RepositoryCollaborator{}, err
	}
	return collaborator, nil
}

func (s *Store) UpdateRepositoryStats(ctx context.Context, owner, name string, sizeBytes int64, indexedAt, maintainedAt *time.Time) error {
	repository, err := s.GetRepositoryByOwnerAndName(ctx, owner, name)
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx, updateRepositoryStatsByIDQuery, repository.ID, sizeBytes, indexedAt, maintainedAt)
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
	repository, err := s.GetRepositoryByOwnerAndName(ctx, owner, name)
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx, deleteRepositoryByIDQuery, repository.ID)
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
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire repository lease connection: %w", err)
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin repository lease transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.ExecContext(ctx, acquireRepositoryLeaseQuery, store.RepositoryLeaseKey(owner, name)); err != nil {
		return fmt.Errorf("acquire repository lease: %w", err)
	}

	if err := fn(ctx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit repository lease transaction: %w", err)
	}
	committed = true
	return nil
}

func (s *Store) Check(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) organizationSlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	if err := s.db.QueryRowContext(ctx, organizationSlugExistsQuery, slug).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) usernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	if err := s.db.QueryRowContext(ctx, usernameExistsQuery, username).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func normalizeOwnerType(value string) (string, error) {
	switch store.NormalizeIdentity(value) {
	case "", store.OwnerTypeUser:
		return store.OwnerTypeUser, nil
	case store.OwnerTypeOrganization:
		return store.OwnerTypeOrganization, nil
	default:
		return "", store.ErrInvalidArgument
	}
}

func normalizeOrganizationRole(value string) (string, error) {
	switch store.NormalizeIdentity(value) {
	case store.OrganizationRoleMember:
		return store.OrganizationRoleMember, nil
	case store.OrganizationRoleMaintainer:
		return store.OrganizationRoleMaintainer, nil
	case store.OrganizationRoleOwner:
		return store.OrganizationRoleOwner, nil
	default:
		return "", store.ErrInvalidArgument
	}
}

func normalizeRepositoryRole(value string) (string, error) {
	switch store.NormalizeIdentity(value) {
	case store.RepositoryRoleRead:
		return store.RepositoryRoleRead, nil
	case store.RepositoryRoleWrite:
		return store.RepositoryRoleWrite, nil
	case store.RepositoryRoleAdmin:
		return store.RepositoryRoleAdmin, nil
	default:
		return "", store.ErrInvalidArgument
	}
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
			&repository.OwnerType,
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
