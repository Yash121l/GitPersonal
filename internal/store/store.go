package store

import (
	"context"
	"errors"
	"hash/fnv"
	"strings"
	"time"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrAlreadyExists   = errors.New("already exists")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrValidation      = errors.New("validation failed")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidArgument = errors.New("invalid argument")
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Repository struct {
	ID               int64      `json:"id"`
	Owner            string     `json:"owner"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Visibility       string     `json:"visibility"`
	DefaultBranch    string     `json:"default_branch"`
	Archived         bool       `json:"archived"`
	RepoPath         string     `json:"-"`
	SizeBytes        int64      `json:"size_bytes"`
	LastIndexedAt    *time.Time `json:"last_indexed_at,omitempty"`
	LastMaintainedAt *time.Time `json:"last_maintained_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreateRepositoryParams struct {
	Owner         string
	Name          string
	Description   string
	Visibility    string
	DefaultBranch string
	RepoPath      string
}

type Session struct {
	ID        int64
	UserID    int64
	TokenID   string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

type CreateSessionParams struct {
	UserID    int64
	TokenID   string
	ExpiresAt time.Time
}

type SSHKey struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"user_id"`
	Name              string     `json:"name"`
	PublicKey         string     `json:"public_key"`
	FingerprintSHA256 string     `json:"fingerprint_sha256"`
	CreatedAt         time.Time  `json:"created_at"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty"`
}

type CreateSSHKeyParams struct {
	UserID            int64
	Name              string
	PublicKey         string
	FingerprintSHA256 string
}

type Store interface {
	CreateUser(ctx context.Context, username, passwordHash, role string) (User, error)
	GetUserByID(ctx context.Context, id int64) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	CreateRepository(ctx context.Context, params CreateRepositoryParams) (Repository, error)
	GetRepositoryByOwnerAndName(ctx context.Context, owner, name string) (Repository, error)
	ListRepositories(ctx context.Context) ([]Repository, error)
	ListRepositoriesByOwner(ctx context.Context, owner string) ([]Repository, error)
	UpdateRepositoryStats(ctx context.Context, owner, name string, sizeBytes int64, indexedAt, maintainedAt *time.Time) error
	DeleteRepository(ctx context.Context, owner, name string) error
	CreateSession(ctx context.Context, params CreateSessionParams) (Session, error)
	GetSessionByTokenID(ctx context.Context, tokenID string) (Session, error)
	RevokeSession(ctx context.Context, tokenID string, revokedAt time.Time) error
	CreateSSHKey(ctx context.Context, params CreateSSHKeyParams) (SSHKey, error)
	GetUserBySSHFingerprint(ctx context.Context, fingerprint string) (User, error)
	TouchSSHKeyUsage(ctx context.Context, fingerprint string, usedAt time.Time) error
	WithRepositoryLease(ctx context.Context, owner, name string, fn func(context.Context) error) error
	Check(ctx context.Context) error
}

func RepositoryLeaseKey(owner, name string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strings.ToLower(strings.TrimSpace(owner))))
	_, _ = hasher.Write([]byte("/"))
	_, _ = hasher.Write([]byte(strings.ToLower(strings.TrimSpace(name))))
	return int64(hasher.Sum64())
}
