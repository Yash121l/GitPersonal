package repository

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yashlunawat/forge/internal/store"
	"github.com/yashlunawat/forge/internal/store/memory"
)

func TestCreateAndDeleteRepositoryProvisioning(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	st := memory.NewStore()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := NewService(logger, st, reposRoot)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

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

func TestMaintenanceUpdatesRepositoryStats(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	st := memory.NewStore()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := NewService(logger, st, reposRoot)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.Start(t.Context())

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

	worktree := t.TempDir()
	runGit(t, worktree, "init")
	runGit(t, worktree, "config", "user.email", "yash@example.com")
	runGit(t, worktree, "config", "user.name", "Yash")
	if err := os.WriteFile(filepath.Join(worktree, "README.md"), []byte("forge\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, worktree, "add", "README.md")
	runGit(t, worktree, "commit", "-m", "seed")
	runGit(t, worktree, "push", repository.RepoPath, "HEAD:refs/heads/main")

	service.ScheduleMaintenance(repository)

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		current, err := st.GetRepositoryByOwnerAndName(context.Background(), "yash", "forge")
		if err != nil {
			t.Fatalf("get repository: %v", err)
		}
		if current.SizeBytes > 0 && current.LastIndexedAt != nil && current.LastMaintainedAt != nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	current, err := st.GetRepositoryByOwnerAndName(context.Background(), "yash", "forge")
	if err != nil {
		t.Fatalf("get repository: %v", err)
	}
	t.Fatalf("maintenance did not update stats: %+v", current)
}

func TestRepoPathUsesFanoutLayout(t *testing.T) {
	t.Parallel()

	provisioner := NewFilesystemProvisioner("/data/repos")
	path := provisioner.RepoPath("Yash", "Forge")

	if filepath.Base(path) != "forge.git" {
		t.Fatalf("unexpected repo basename: %s", path)
	}
	relative, err := filepath.Rel("/data/repos", path)
	if err != nil {
		t.Fatalf("relative repo path: %v", err)
	}
	parts := strings.Split(relative, string(filepath.Separator))
	if len(parts) != 4 {
		t.Fatalf("expected sharded path with 4 parts, got %v", parts)
	}
	if len(parts[0]) != 2 || len(parts[1]) != 2 {
		t.Fatalf("expected 2-byte shard prefixes, got %v", parts[:2])
	}
}

func TestPermissionsCoverOrganizationsAndCollaborators(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	st := memory.NewStore()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := NewService(logger, st, reposRoot)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	alice, err := st.CreateUser(context.Background(), "alice", "hash", "member")
	if err != nil {
		t.Fatalf("create alice: %v", err)
	}
	bob, err := st.CreateUser(context.Background(), "bob", "hash", "member")
	if err != nil {
		t.Fatalf("create bob: %v", err)
	}
	carol, err := st.CreateUser(context.Background(), "carol", "hash", "member")
	if err != nil {
		t.Fatalf("create carol: %v", err)
	}

	if _, err := st.CreateOrganization(context.Background(), store.CreateOrganizationParams{
		Slug:        "team",
		DisplayName: "Team",
		Description: "shared ownership",
		CreatedBy:   alice.ID,
	}); err != nil {
		t.Fatalf("create organization: %v", err)
	}
	if _, err := st.AddOrganizationMember(context.Background(), store.AddOrganizationMemberParams{
		OrganizationSlug: "team",
		Username:         "bob",
		Role:             store.OrganizationRoleMaintainer,
	}); err != nil {
		t.Fatalf("add bob to org: %v", err)
	}

	orgRepo, err := service.CreateRepository(context.Background(), store.CreateRepositoryParams{
		Owner:         "team",
		OwnerType:     store.OwnerTypeOrganization,
		Name:          "infra",
		Description:   "org repo",
		Visibility:    "private",
		DefaultBranch: "main",
	})
	if err != nil {
		t.Fatalf("create org repository: %v", err)
	}

	if _, err := st.AddRepositoryCollaborator(context.Background(), store.AddRepositoryCollaboratorParams{
		Owner:    "team",
		RepoName: "infra",
		Username: "carol",
		Role:     store.RepositoryRoleRead,
	}); err != nil {
		t.Fatalf("add carol collaborator: %v", err)
	}

	alicePermissions, err := service.Permissions(context.Background(), &alice, orgRepo)
	if err != nil {
		t.Fatalf("alice permissions: %v", err)
	}
	if !alicePermissions.CanAdmin || !alicePermissions.CanWrite || !alicePermissions.CanRead {
		t.Fatalf("expected alice to admin org repo, got %+v", alicePermissions)
	}

	bobPermissions, err := service.Permissions(context.Background(), &bob, orgRepo)
	if err != nil {
		t.Fatalf("bob permissions: %v", err)
	}
	if !bobPermissions.CanRead || !bobPermissions.CanWrite || bobPermissions.CanAdmin {
		t.Fatalf("expected bob maintainer permissions, got %+v", bobPermissions)
	}

	carolPermissions, err := service.Permissions(context.Background(), &carol, orgRepo)
	if err != nil {
		t.Fatalf("carol permissions: %v", err)
	}
	if !carolPermissions.CanRead || carolPermissions.CanWrite || carolPermissions.CanAdmin {
		t.Fatalf("expected carol read-only collaborator permissions, got %+v", carolPermissions)
	}

	anonymousPermissions, err := service.Permissions(context.Background(), nil, orgRepo)
	if err != nil {
		t.Fatalf("anonymous permissions: %v", err)
	}
	if anonymousPermissions.CanRead || anonymousPermissions.CanWrite || anonymousPermissions.CanAdmin {
		t.Fatalf("expected private repo to block anonymous access, got %+v", anonymousPermissions)
	}

	userRepo, err := service.CreateRepository(context.Background(), store.CreateRepositoryParams{
		Owner:         "alice",
		OwnerType:     store.OwnerTypeUser,
		Name:          "personal",
		Description:   "user repo",
		Visibility:    "private",
		DefaultBranch: "main",
	})
	if err != nil {
		t.Fatalf("create user repository: %v", err)
	}

	if _, err := st.AddRepositoryCollaborator(context.Background(), store.AddRepositoryCollaboratorParams{
		Owner:    "alice",
		RepoName: "personal",
		Username: "bob",
		Role:     store.RepositoryRoleAdmin,
	}); err != nil {
		t.Fatalf("add bob admin collaborator: %v", err)
	}

	userRepoPermissions, err := service.Permissions(context.Background(), &bob, userRepo)
	if err != nil {
		t.Fatalf("bob user repo permissions: %v", err)
	}
	if !userRepoPermissions.CanAdmin || !userRepoPermissions.CanWrite || !userRepoPermissions.CanRead {
		t.Fatalf("expected bob admin collaborator permissions, got %+v", userRepoPermissions)
	}
}

func runGit(t *testing.T, workdir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = workdir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}
