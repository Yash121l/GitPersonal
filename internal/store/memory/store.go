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
	usersByID map[int64]store.User
	repos     map[string]store.Repository
	userRepos map[string][]string
	sessions  map[string]store.Session
	sshKeys   map[string]store.SSHKey

	lockMu    sync.Mutex
	repoLocks map[string]*sync.Mutex
}

func NewStore() *Store {
	return &Store{
		nextID:    1,
		users:     make(map[string]store.User),
		usersByID: make(map[int64]store.User),
		repos:     make(map[string]store.Repository),
		userRepos: make(map[string][]string),
		sessions:  make(map[string]store.Session),
		sshKeys:   make(map[string]store.SSHKey),
		repoLocks: make(map[string]*sync.Mutex),
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
	s.usersByID[user.ID] = user

	return user, nil
}

func (s *Store) GetUserByID(_ context.Context, id int64) (store.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.usersByID[id]
	if !ok {
		return store.User{}, store.ErrNotFound
	}

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
		ID:               s.next(),
		Owner:            params.Owner,
		Name:             params.Name,
		Description:      params.Description,
		Visibility:       params.Visibility,
		DefaultBranch:    params.DefaultBranch,
		RepoPath:         params.RepoPath,
		CreatedAt:        now,
		UpdatedAt:        now,
		LastIndexedAt:    &now,
		LastMaintainedAt: &now,
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

func (s *Store) ListRepositories(_ context.Context) ([]store.Repository, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repositories := make([]store.Repository, 0, len(s.repos))
	for _, repository := range s.repos {
		repositories = append(repositories, repository)
	}

	sort.Slice(repositories, func(i, j int) bool {
		if repositories[i].Owner == repositories[j].Owner {
			return repositories[i].Name < repositories[j].Name
		}
		return repositories[i].Owner < repositories[j].Owner
	})

	return repositories, nil
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

func (s *Store) UpdateRepositoryStats(_ context.Context, owner, name string, sizeBytes int64, indexedAt, maintainedAt *time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := repoKey(owner, name)
	repository, ok := s.repos[key]
	if !ok {
		return store.ErrNotFound
	}

	repository.SizeBytes = sizeBytes
	if indexedAt != nil {
		value := *indexedAt
		repository.LastIndexedAt = &value
	}
	if maintainedAt != nil {
		value := *maintainedAt
		repository.LastMaintainedAt = &value
	}
	repository.UpdatedAt = time.Now().UTC()
	s.repos[key] = repository
	return nil
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

func (s *Store) CreateSession(_ context.Context, params store.CreateSessionParams) (store.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[params.TokenID]; exists {
		return store.Session{}, store.ErrAlreadyExists
	}

	session := store.Session{
		ID:        s.next(),
		UserID:    params.UserID,
		TokenID:   params.TokenID,
		ExpiresAt: params.ExpiresAt,
		CreatedAt: time.Now().UTC(),
	}
	s.sessions[params.TokenID] = session
	return session, nil
}

func (s *Store) GetSessionByTokenID(_ context.Context, tokenID string) (store.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[tokenID]
	if !ok {
		return store.Session{}, store.ErrNotFound
	}
	return session, nil
}

func (s *Store) RevokeSession(_ context.Context, tokenID string, revokedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[tokenID]
	if !ok {
		return store.ErrNotFound
	}
	value := revokedAt
	session.RevokedAt = &value
	s.sessions[tokenID] = session
	return nil
}

func (s *Store) CreateSSHKey(_ context.Context, params store.CreateSSHKeyParams) (store.SSHKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sshKeys[params.FingerprintSHA256]; exists {
		return store.SSHKey{}, store.ErrAlreadyExists
	}

	key := store.SSHKey{
		ID:                s.next(),
		UserID:            params.UserID,
		Name:              params.Name,
		PublicKey:         params.PublicKey,
		FingerprintSHA256: params.FingerprintSHA256,
		CreatedAt:         time.Now().UTC(),
	}
	s.sshKeys[params.FingerprintSHA256] = key
	return key, nil
}

func (s *Store) GetUserBySSHFingerprint(_ context.Context, fingerprint string) (store.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, ok := s.sshKeys[fingerprint]
	if !ok {
		return store.User{}, store.ErrNotFound
	}
	user, ok := s.usersByID[key.UserID]
	if !ok {
		return store.User{}, store.ErrNotFound
	}
	return user, nil
}

func (s *Store) TouchSSHKeyUsage(_ context.Context, fingerprint string, usedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, ok := s.sshKeys[fingerprint]
	if !ok {
		return store.ErrNotFound
	}
	value := usedAt
	key.LastUsedAt = &value
	s.sshKeys[fingerprint] = key
	return nil
}

func (s *Store) WithRepositoryLease(ctx context.Context, owner, name string, fn func(context.Context) error) error {
	mutex := s.repoMutex(owner, name)
	mutex.Lock()
	defer mutex.Unlock()

	return fn(ctx)
}

func (s *Store) Check(_ context.Context) error {
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

func (s *Store) repoMutex(owner, name string) *sync.Mutex {
	key := repoKey(owner, name)

	s.lockMu.Lock()
	defer s.lockMu.Unlock()

	mutex, ok := s.repoLocks[key]
	if ok {
		return mutex
	}

	mutex = &sync.Mutex{}
	s.repoLocks[key] = mutex
	return mutex
}
