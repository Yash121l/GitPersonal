package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yashlunawat/forge/internal/store"
)

type Store struct {
	mu        sync.RWMutex
	nextID    int64
	users     map[string]store.User
	repos     map[string]store.Repository
	userRepos map[string][]string
}

func NewStore() *Store {
	return &Store{
		nextID:    1,
		users:     make(map[string]store.User),
		repos:     make(map[string]store.Repository),
		userRepos: make(map[string][]string),
	}
}

func (s *Store) CreateUser(_ context.Context, username, passwordHash, role string) (store.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := normalize(username)
	if _, exists := s.users[key]; exists {
		return store.User{}, store.ErrAlreadyExists
	}

	now := time.Now().UTC()
	user := store.User{
		ID:           s.next(),
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
	}
	s.users[key] = user

	return user, nil
}

func (s *Store) GetUserByUsername(_ context.Context, username string) (store.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[normalize(username)]
	if !ok {
		return store.User{}, store.ErrNotFound
	}

	return user, nil
}

func (s *Store) CreateRepository(_ context.Context, params store.CreateRepositoryParams) (store.Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := repoKey(params.Owner, params.Name)
	if _, exists := s.repos[key]; exists {
		return store.Repository{}, store.ErrAlreadyExists
	}

	now := time.Now().UTC()
	repo := store.Repository{
		ID:            s.next(),
		Owner:         params.Owner,
		Name:          params.Name,
		Description:   params.Description,
		Visibility:    params.Visibility,
		DefaultBranch: params.DefaultBranch,
		RepoPath:      params.RepoPath,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	s.repos[key] = repo
	s.userRepos[normalize(params.Owner)] = append(s.userRepos[normalize(params.Owner)], key)

	return repo, nil
}

func (s *Store) GetRepositoryByOwnerAndName(_ context.Context, owner, name string) (store.Repository, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repository, ok := s.repos[repoKey(owner, name)]
	if !ok {
		return store.Repository{}, store.ErrNotFound
	}

	return repository, nil
}

func (s *Store) ListRepositoriesByOwner(_ context.Context, owner string) ([]store.Repository, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := append([]string(nil), s.userRepos[normalize(owner)]...)
	repositories := make([]store.Repository, 0, len(keys))
	for _, key := range keys {
		repository, ok := s.repos[key]
		if ok {
			repositories = append(repositories, repository)
		}
	}

	sort.Slice(repositories, func(i, j int) bool {
		return repositories[i].Name < repositories[j].Name
	})

	return repositories, nil
}

func (s *Store) DeleteRepository(_ context.Context, owner, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := repoKey(owner, name)
	if _, ok := s.repos[key]; !ok {
		return store.ErrNotFound
	}

	delete(s.repos, key)

	ownerKey := normalize(owner)
	keys := s.userRepos[ownerKey]
	filtered := keys[:0]
	for _, existing := range keys {
		if existing != key {
			filtered = append(filtered, existing)
		}
	}
	s.userRepos[ownerKey] = filtered

	return nil
}

func (s *Store) next() int64 {
	id := s.nextID
	s.nextID++
	return id
}

func repoKey(owner, name string) string {
	return normalize(owner) + "/" + normalize(name)
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
