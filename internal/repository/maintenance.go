package repository

import (
	"context"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/yashlunawat/forge/internal/store"
)

type MaintenanceScheduler struct {
	logger      *slog.Logger
	store       store.Store
	provisioner *FilesystemProvisioner

	queue   chan store.Repository
	pending sync.Map
}

func NewMaintenanceScheduler(logger *slog.Logger, st store.Store, provisioner *FilesystemProvisioner) *MaintenanceScheduler {
	return &MaintenanceScheduler{
		logger:      logger,
		store:       st,
		provisioner: provisioner,
		queue:       make(chan store.Repository, 256),
	}
}

func (m *MaintenanceScheduler) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case repository := <-m.queue:
				m.pending.Delete(repoPendingKey(repository))
				m.process(ctx, repository)
			}
		}
	}()
}

func (m *MaintenanceScheduler) Enqueue(repository store.Repository) {
	key := repoPendingKey(repository)
	if _, loaded := m.pending.LoadOrStore(key, struct{}{}); loaded {
		return
	}

	select {
	case m.queue <- repository:
	default:
		m.pending.Delete(key)
		m.logger.Warn("maintenance queue full", "owner", repository.Owner, "repo", repository.Name)
	}
}

func (m *MaintenanceScheduler) process(ctx context.Context, repository store.Repository) {
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	indexedAt := time.Now().UTC()
	sizeBytes, err := repositorySize(repository.RepoPath)
	if err != nil {
		m.logger.Error("index repository size", "owner", repository.Owner, "repo", repository.Name, "error", err)
		return
	}

	if _, err := m.provisioner.RunGit(runCtx, repository.RepoPath, "gc", "--auto"); err != nil {
		m.logger.Warn("git gc --auto failed", "owner", repository.Owner, "repo", repository.Name, "error", err)
	}
	if _, err := m.provisioner.RunGit(runCtx, repository.RepoPath, "commit-graph", "write", "--reachable"); err != nil {
		m.logger.Warn("git commit-graph write failed", "owner", repository.Owner, "repo", repository.Name, "error", err)
	}

	maintainedAt := time.Now().UTC()
	if err := m.store.UpdateRepositoryStats(runCtx, repository.Owner, repository.Name, sizeBytes, &indexedAt, &maintainedAt); err != nil {
		m.logger.Error("update repository stats", "owner", repository.Owner, "repo", repository.Name, "error", err)
	}
}

func repoPendingKey(repository store.Repository) string {
	return repository.Owner + "/" + repository.Name
}

func repositorySize(root string) (int64, error) {
	var total int64
	err := filepath.WalkDir(root, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	if err != nil {
		return 0, err
	}
	return total, nil
}
