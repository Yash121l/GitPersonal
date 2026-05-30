package repository

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/yashlunawat/forge/internal/store"
)

type Permissions struct {
	CanRead  bool
	CanWrite bool
	CanAdmin bool
}

type Service struct {
	logger      *slog.Logger
	store       store.Store
	provisioner *FilesystemProvisioner
	maintenance *MaintenanceScheduler
}

func NewService(logger *slog.Logger, st store.Store, reposRoot string) (*Service, error) {
	provisioner := NewFilesystemProvisioner(reposRoot)
	service := &Service{
		logger:      logger,
		store:       st,
		provisioner: provisioner,
		maintenance: NewMaintenanceScheduler(logger, st, provisioner),
	}
	if err := service.Check(context.Background()); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *Service) CreateRepository(ctx context.Context, params store.CreateRepositoryParams) (store.Repository, error) {
	var repository store.Repository
	err := s.store.WithRepositoryLease(ctx, params.Owner, params.Name, func(ctx context.Context) error {
		repoPath, err := s.provisioner.CreateBareRepository(ctx, params.Owner, params.Name, params.DefaultBranch)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				return store.ErrAlreadyExists
			}
			return err
		}

		params.RepoPath = repoPath
		repository, err = s.store.CreateRepository(ctx, params)
		if err != nil {
			cleanupErr := s.provisioner.CleanupRepository(repoPath)
			if cleanupErr != nil {
				s.logger.Error("cleanup repo after metadata failure", "repo_path", repoPath, "error", cleanupErr)
			}
			return err
		}
		s.maintenance.Enqueue(repository)
		return nil
	})
	if err != nil {
		return store.Repository{}, err
	}

	return repository, nil
}

func (s *Service) DeleteRepository(ctx context.Context, owner, name string) error {
	return s.store.WithRepositoryLease(ctx, owner, name, func(ctx context.Context) error {
		repository, err := s.store.GetRepositoryByOwnerAndName(ctx, owner, name)
		if err != nil {
			return err
		}

		trashPath, restore, err := s.provisioner.StageDelete(repository.RepoPath)
		if err != nil {
			return err
		}

		if err := s.store.DeleteRepository(ctx, owner, name); err != nil {
			if restore != nil {
				if restoreErr := restore(); restoreErr != nil {
					s.logger.Error("restore repo after metadata delete failure", "repo_path", repository.RepoPath, "error", restoreErr)
				}
			}
			return err
		}

		if trashPath != "" {
			if err := s.provisioner.CleanupRepository(trashPath); err != nil {
				s.logger.Error("cleanup trashed repo", "repo_path", trashPath, "error", err)
			}
		}

		return nil
	})
}

func (s *Service) Check(ctx context.Context) error {
	if err := s.store.Check(ctx); err != nil {
		return err
	}
	return s.provisioner.Check(ctx)
}

func (s *Service) Start(ctx context.Context) {
	s.maintenance.Start(ctx)
}

func (s *Service) GetRepository(ctx context.Context, owner, name string) (store.Repository, error) {
	return s.store.GetRepositoryByOwnerAndName(ctx, owner, name)
}

func (s *Service) RelativeRepoPath(repository store.Repository) (string, error) {
	return s.provisioner.RelativeRepoPath(repository.RepoPath)
}

func (s *Service) Permissions(ctx context.Context, user *store.User, repository store.Repository) (Permissions, error) {
	permissions := Permissions{}

	if repository.Visibility == "public" {
		permissions.CanRead = true
	}
	if user == nil {
		return permissions, nil
	}
	if user.Role == store.OrganizationRoleOwner {
		return Permissions{CanRead: true, CanWrite: true, CanAdmin: true}, nil
	}

	if repository.OwnerType == store.OwnerTypeUser && strings.EqualFold(user.Username, repository.Owner) {
		return Permissions{CanRead: true, CanWrite: true, CanAdmin: true}, nil
	}

	if repository.OwnerType == store.OwnerTypeOrganization {
		membership, err := s.store.GetOrganizationMembership(ctx, repository.Owner, user.ID)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return Permissions{}, err
		}
		switch membership.Role {
		case store.OrganizationRoleOwner:
			permissions.CanRead = true
			permissions.CanWrite = true
			permissions.CanAdmin = true
		case store.OrganizationRoleMaintainer:
			permissions.CanRead = true
			permissions.CanWrite = true
		case store.OrganizationRoleMember:
			permissions.CanRead = true
		}
	}

	collaborator, err := s.store.GetRepositoryCollaborator(ctx, repository.Owner, repository.Name, user.ID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return Permissions{}, err
	}
	switch collaborator.Role {
	case store.RepositoryRoleAdmin:
		permissions.CanRead = true
		permissions.CanWrite = true
		permissions.CanAdmin = true
	case store.RepositoryRoleWrite:
		permissions.CanRead = true
		permissions.CanWrite = true
	case store.RepositoryRoleRead:
		permissions.CanRead = true
	}

	return permissions, nil
}

func (s *Service) CanRead(ctx context.Context, user *store.User, repository store.Repository) (bool, error) {
	permissions, err := s.Permissions(ctx, user, repository)
	if err != nil {
		return false, err
	}
	return permissions.CanRead, nil
}

func (s *Service) CanWrite(ctx context.Context, user *store.User, repository store.Repository) (bool, error) {
	permissions, err := s.Permissions(ctx, user, repository)
	if err != nil {
		return false, err
	}
	return permissions.CanWrite, nil
}

func (s *Service) CanAdmin(ctx context.Context, user *store.User, repository store.Repository) (bool, error) {
	permissions, err := s.Permissions(ctx, user, repository)
	if err != nil {
		return false, err
	}
	return permissions.CanAdmin, nil
}

func (s *Service) ScheduleMaintenance(repository store.Repository) {
	s.maintenance.Enqueue(repository)
}
