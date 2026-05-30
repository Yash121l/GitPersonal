package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/yashlunawat/forge/internal/store"
)

type Store struct {
	mu                  sync.RWMutex
	nextID              int64
	users               map[string]store.User
	usersByID           map[int64]store.User
	organizations       map[string]store.Organization
	organizationMembers map[string]map[int64]store.OrganizationMembership
	repos               map[string]store.Repository
	ownerRepos          map[string][]string
	repoCollaborators   map[string]map[int64]store.RepositoryCollaborator
	sessions            map[string]store.Session
	sshKeys             map[string]store.SSHKey

	lockMu    sync.Mutex
	repoLocks map[string]*sync.Mutex
}

func NewStore() *Store {
	return &Store{
		nextID:              1,
		users:               make(map[string]store.User),
		usersByID:           make(map[int64]store.User),
		organizations:       make(map[string]store.Organization),
		organizationMembers: make(map[string]map[int64]store.OrganizationMembership),
		repos:               make(map[string]store.Repository),
		ownerRepos:          make(map[string][]string),
		repoCollaborators:   make(map[string]map[int64]store.RepositoryCollaborator),
		sessions:            make(map[string]store.Session),
		sshKeys:             make(map[string]store.SSHKey),
		repoLocks:           make(map[string]*sync.Mutex),
	}
}

func (s *Store) CreateUser(_ context.Context, username, passwordHash, role string) (store.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := normalize(username)
	if _, exists := s.users[key]; exists {
		return store.User{}, store.ErrAlreadyExists
	}
	if _, exists := s.organizations[key]; exists {
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

func (s *Store) CreateOrganization(_ context.Context, params store.CreateOrganizationParams) (store.Organization, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := normalize(params.Slug)
	if _, exists := s.organizations[key]; exists {
		return store.Organization{}, store.ErrAlreadyExists
	}
	if _, exists := s.users[key]; exists {
		return store.Organization{}, store.ErrAlreadyExists
	}

	creator, ok := s.usersByID[params.CreatedBy]
	if !ok {
		return store.Organization{}, store.ErrNotFound
	}

	now := time.Now().UTC()
	org := store.Organization{
		ID:          s.next(),
		Slug:        params.Slug,
		DisplayName: params.DisplayName,
		Description: params.Description,
		CreatedBy:   params.CreatedBy,
		CreatedAt:   now,
	}
	s.organizations[key] = org

	if _, ok := s.organizationMembers[key]; !ok {
		s.organizationMembers[key] = make(map[int64]store.OrganizationMembership)
	}
	s.organizationMembers[key][creator.ID] = store.OrganizationMembership{
		OrganizationID:          org.ID,
		OrganizationSlug:        org.Slug,
		OrganizationDisplayName: org.DisplayName,
		UserID:                  creator.ID,
		Username:                creator.Username,
		Role:                    store.OrganizationRoleOwner,
		CreatedAt:               now,
	}

	return org, nil
}

func (s *Store) GetOrganizationBySlug(_ context.Context, slug string) (store.Organization, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.organizations[normalize(slug)]
	if !ok {
		return store.Organization{}, store.ErrNotFound
	}

	return org, nil
}

func (s *Store) ListOrganizationsByMember(_ context.Context, userID int64) ([]store.OrganizationMembership, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memberships := make([]store.OrganizationMembership, 0)
	for _, members := range s.organizationMembers {
		if membership, ok := members[userID]; ok {
			memberships = append(memberships, membership)
		}
	}

	sort.Slice(memberships, func(i, j int) bool {
		return normalize(memberships[i].OrganizationSlug) < normalize(memberships[j].OrganizationSlug)
	})
	return memberships, nil
}

func (s *Store) AddOrganizationMember(_ context.Context, params store.AddOrganizationMemberParams) (store.OrganizationMembership, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, err := normalizeOrganizationRole(params.Role)
	if err != nil {
		return store.OrganizationMembership{}, err
	}

	orgKey := normalize(params.OrganizationSlug)
	org, ok := s.organizations[orgKey]
	if !ok {
		return store.OrganizationMembership{}, store.ErrNotFound
	}

	user, ok := s.users[normalize(params.Username)]
	if !ok {
		return store.OrganizationMembership{}, store.ErrNotFound
	}

	if _, ok := s.organizationMembers[orgKey]; !ok {
		s.organizationMembers[orgKey] = make(map[int64]store.OrganizationMembership)
	}
	if _, exists := s.organizationMembers[orgKey][user.ID]; exists {
		return store.OrganizationMembership{}, store.ErrAlreadyExists
	}

	membership := store.OrganizationMembership{
		OrganizationID:          org.ID,
		OrganizationSlug:        org.Slug,
		OrganizationDisplayName: org.DisplayName,
		UserID:                  user.ID,
		Username:                user.Username,
		Role:                    role,
		CreatedAt:               time.Now().UTC(),
	}
	s.organizationMembers[orgKey][user.ID] = membership

	return membership, nil
}

func (s *Store) GetOrganizationMembership(_ context.Context, organizationSlug string, userID int64) (store.OrganizationMembership, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	members, ok := s.organizationMembers[normalize(organizationSlug)]
	if !ok {
		return store.OrganizationMembership{}, store.ErrNotFound
	}
	membership, ok := members[userID]
	if !ok {
		return store.OrganizationMembership{}, store.ErrNotFound
	}
	return membership, nil
}

func (s *Store) CreateRepository(_ context.Context, params store.CreateRepositoryParams) (store.Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ownerType, err := normalizeOwnerType(params.OwnerType)
	if err != nil {
		return store.Repository{}, err
	}

	switch ownerType {
	case store.OwnerTypeUser:
		if _, ok := s.users[normalize(params.Owner)]; !ok {
			return store.Repository{}, store.ErrNotFound
		}
	case store.OwnerTypeOrganization:
		if _, ok := s.organizations[normalize(params.Owner)]; !ok {
			return store.Repository{}, store.ErrNotFound
		}
	}

	key := repoKey(params.Owner, params.Name)
	if _, exists := s.repos[key]; exists {
		return store.Repository{}, store.ErrAlreadyExists
	}

	now := time.Now().UTC()
	repo := store.Repository{
		ID:               s.next(),
		Owner:            params.Owner,
		OwnerType:        ownerType,
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
	ownerKey := normalize(params.Owner)
	s.ownerRepos[ownerKey] = append(s.ownerRepos[ownerKey], key)

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
	sortRepositories(repositories)
	return repositories, nil
}

func (s *Store) ListRepositoriesByOwner(_ context.Context, owner string) ([]store.Repository, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := append([]string(nil), s.ownerRepos[normalize(owner)]...)
	repositories := make([]store.Repository, 0, len(keys))
	for _, key := range keys {
		if repository, ok := s.repos[key]; ok {
			repositories = append(repositories, repository)
		}
	}
	sortRepositories(repositories)
	return repositories, nil
}

func (s *Store) ListRepositoriesForUser(_ context.Context, userID int64) ([]store.Repository, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.usersByID[userID]
	if !ok {
		return nil, store.ErrNotFound
	}

	repositories := make([]store.Repository, 0)
	for key, repository := range s.repos {
		if repository.OwnerType == store.OwnerTypeUser && normalize(repository.Owner) == normalize(user.Username) {
			repositories = append(repositories, repository)
			continue
		}
		if repository.OwnerType == store.OwnerTypeOrganization {
			if members, ok := s.organizationMembers[normalize(repository.Owner)]; ok {
				if _, ok := members[userID]; ok {
					repositories = append(repositories, repository)
					continue
				}
			}
		}
		if collaborators, ok := s.repoCollaborators[key]; ok {
			if _, ok := collaborators[userID]; ok {
				repositories = append(repositories, repository)
			}
		}
	}

	sortRepositories(repositories)
	return repositories, nil
}

func (s *Store) AddRepositoryCollaborator(_ context.Context, params store.AddRepositoryCollaboratorParams) (store.RepositoryCollaborator, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, err := normalizeRepositoryRole(params.Role)
	if err != nil {
		return store.RepositoryCollaborator{}, err
	}

	user, ok := s.users[normalize(params.Username)]
	if !ok {
		return store.RepositoryCollaborator{}, store.ErrNotFound
	}

	repository, ok := s.repos[repoKey(params.Owner, params.RepoName)]
	if !ok {
		return store.RepositoryCollaborator{}, store.ErrNotFound
	}

	key := repoKey(repository.Owner, repository.Name)
	if _, ok := s.repoCollaborators[key]; !ok {
		s.repoCollaborators[key] = make(map[int64]store.RepositoryCollaborator)
	}
	if _, exists := s.repoCollaborators[key][user.ID]; exists {
		return store.RepositoryCollaborator{}, store.ErrAlreadyExists
	}

	collaborator := store.RepositoryCollaborator{
		RepositoryID: repository.ID,
		UserID:       user.ID,
		Username:     user.Username,
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}
	s.repoCollaborators[key][user.ID] = collaborator
	return collaborator, nil
}

func (s *Store) GetRepositoryCollaborator(_ context.Context, owner, repoName string, userID int64) (store.RepositoryCollaborator, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	collaborators, ok := s.repoCollaborators[repoKey(owner, repoName)]
	if !ok {
		return store.RepositoryCollaborator{}, store.ErrNotFound
	}
	collaborator, ok := collaborators[userID]
	if !ok {
		return store.RepositoryCollaborator{}, store.ErrNotFound
	}
	return collaborator, nil
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
	delete(s.repoCollaborators, key)

	ownerKey := normalize(owner)
	keys := s.ownerRepos[ownerKey]
	filtered := keys[:0]
	for _, existing := range keys {
		if existing != key {
			filtered = append(filtered, existing)
		}
	}
	s.ownerRepos[ownerKey] = filtered

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
	return store.NormalizeIdentity(value)
}

func sortRepositories(repositories []store.Repository) {
	sort.Slice(repositories, func(i, j int) bool {
		if repositories[i].Owner == repositories[j].Owner {
			if repositories[i].OwnerType == repositories[j].OwnerType {
				return repositories[i].Name < repositories[j].Name
			}
			return repositories[i].OwnerType < repositories[j].OwnerType
		}
		return repositories[i].Owner < repositories[j].Owner
	})
}

func normalizeOwnerType(value string) (string, error) {
	switch normalize(value) {
	case "", store.OwnerTypeUser:
		return store.OwnerTypeUser, nil
	case store.OwnerTypeOrganization:
		return store.OwnerTypeOrganization, nil
	default:
		return "", store.ErrInvalidArgument
	}
}

func normalizeOrganizationRole(value string) (string, error) {
	switch normalize(value) {
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
	switch normalize(value) {
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
