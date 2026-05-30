package repository

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/yashlunawat/forge/internal/store"
	"github.com/yashlunawat/forge/internal/store/memory"
)

func TestCreateAndDeleteRepositoryProvisioning(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	st := memory.NewStore()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := NewService(logger, st, reposRoot)

	if _, err := st.CreateUser(context.Background(), "yash", "hash", "member"); err != nil {
		t.Fatalf("create user: %v", err)
	}

	repository, err := service.CreateRepository(context.Background(), store.CreateRepositoryParams{
		Owner:         "yash",
		Name:          "forge",
		Description:   "Self-hosted git platform",
		Visibility:    "private",
		DefaultBranch: "main",
	})
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	if repository.RepoPath == "" {
		t.Fatal("expected repo path to be set")
	}
	if _, err := os.Stat(repository.RepoPath); err != nil {
		t.Fatalf("stat provisioned repo: %v", err)
	}

	headContents, err := os.ReadFile(filepath.Join(repository.RepoPath, "HEAD"))
	if err != nil {
		t.Fatalf("read HEAD: %v", err)
	}
	if string(headContents) != "ref: refs/heads/main\n" {
		t.Fatalf("unexpected HEAD contents: %q", string(headContents))
	}

	if err := service.DeleteRepository(context.Background(), "yash", "forge"); err != nil {
		t.Fatalf("delete repository: %v", err)
	}
	if _, err := os.Stat(repository.RepoPath); !os.IsNotExist(err) {
		t.Fatalf("expected repo path to be removed, stat err = %v", err)
	}
}
