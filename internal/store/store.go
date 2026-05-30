package store

import (
	"context"
	"errors"
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
	ID            int64     `json:"id"`
	Owner         string    `json:"owner"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Visibility    string    `json:"visibility"`
	DefaultBranch string    `json:"default_branch"`
	Archived      bool      `json:"archived"`
	RepoPath      string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateRepositoryParams struct {
	Owner         string
	Name          string
	Description   string
	Visibility    string
	DefaultBranch string
	RepoPath      string
}

type Store interface {
	CreateUser(ctx context.Context, username, passwordHash, role string) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	CreateRepository(ctx context.Context, params CreateRepositoryParams) (Repository, error)
	GetRepositoryByOwnerAndName(ctx context.Context, owner, name string) (Repository, error)
	ListRepositoriesByOwner(ctx context.Context, owner string) ([]Repository, error)
	DeleteRepository(ctx context.Context, owner, name string) error
}
