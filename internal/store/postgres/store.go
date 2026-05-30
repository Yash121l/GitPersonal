package postgres

import (
	"context"
	"database/sql"
	"errors"

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

	createRepositoryQuery = `
INSERT INTO repositories (
	owner_user_id,
	name,
	description,
	visibility,
	default_branch,
	repo_path
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, description, visibility, default_branch, is_archived, repo_path, created_at, updated_at`

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
	r.created_at,
	r.updated_at
FROM repositories r
JOIN users u ON u.id = r.owner_user_id
WHERE lower(u.username) = lower($1)
ORDER BY r.name ASC`

	deleteRepositoryQuery = `
DELETE FROM repositories r
USING users u
WHERE r.owner_user_id = u.id
  AND lower(u.username) = lower($1)
  AND lower(r.name) = lower($2)`

	repositoryLeaseQuery = `SELECT pg_advisory_xact_lock($1)`
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

func (s *Store) ListRepositoriesByOwner(ctx context.Context, owner string) ([]store.Repository, error) {
	rows, err := s.db.QueryContext(ctx, listRepositoriesByOwnerQuery, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (s *Store) WithRepositoryLease(ctx context.Context, owner, name string, fn func(context.Context) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, repositoryLeaseQuery, store.RepositoryLeaseKey(owner, name)); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := fn(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *Store) Check(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
