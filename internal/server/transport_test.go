package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cryptossh "golang.org/x/crypto/ssh"

	"github.com/yashlunawat/forge/internal/auth"
	"github.com/yashlunawat/forge/internal/config"
	"github.com/yashlunawat/forge/internal/repository"
	"github.com/yashlunawat/forge/internal/sshgit"
	"github.com/yashlunawat/forge/internal/store"
	"github.com/yashlunawat/forge/internal/store/memory"
)

func TestLogoutRevokesSession(t *testing.T) {
	t.Parallel()

	app, _ := newTestServer(t)

	register := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "yash",
		"password": "supersecretpass",
	}, nil)
	cookie := firstCookie(t, register.Result().Cookies(), "forge_session")

	logout := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/auth/logout", nil, cookie)
	if logout.Code != http.StatusOK {
		t.Fatalf("logout status = %d", logout.Code)
	}

	me := performJSONRequest(t, app.Router(), http.MethodGet, "/api/v1/me", nil, cookie)
	if me.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked session to be unauthorized, got %d", me.Code)
	}
}

func TestGitSmartHTTPPushAndLsRemote(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	app, repositories := newTestServer(t)
	passwordHash, err := auth.HashPassword("supersecretpass")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if _, err := app.store.CreateUser(context.Background(), "yash", passwordHash, "member"); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := repositories.CreateRepository(context.Background(), store.CreateRepositoryParams{
		Owner:         "yash",
		Name:          "forge",
		Visibility:    "private",
		DefaultBranch: "main",
	}); err != nil {
		t.Fatalf("create repository: %v", err)
	}

	server := httptest.NewServer(app.Router())
	defer server.Close()

	worktree := t.TempDir()
	runGit(t, worktree, nil, "init")
	runGit(t, worktree, nil, "config", "user.email", "yash@example.com")
	runGit(t, worktree, nil, "config", "user.name", "Yash")
	if err := os.WriteFile(filepath.Join(worktree, "README.md"), []byte("forge\n"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	runGit(t, worktree, nil, "add", "README.md")
	runGit(t, worktree, nil, "commit", "-m", "initial")

	remote := strings.Replace(server.URL, "http://", "http://yash:supersecretpass@", 1) + "/git/yash/forge.git"
	runGit(t, worktree, []string{"GIT_TERMINAL_PROMPT=0"}, "remote", "add", "origin", remote)
	runGit(t, worktree, []string{"GIT_TERMINAL_PROMPT=0"}, "push", "origin", "HEAD:refs/heads/main")

	output := runGit(t, "", []string{"GIT_TERMINAL_PROMPT=0"}, "ls-remote", remote)
	if !strings.Contains(output, "refs/heads/main") {
		t.Fatalf("expected ls-remote to include main branch, output = %s", output)
	}
}

func TestGitSSHUploadPack(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	if _, err := exec.LookPath("ssh"); err != nil {
		t.Skip("ssh not available")
	}

	app, repositories := newTestServer(t)
	passwordHash, err := auth.HashPassword("supersecretpass")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user, err := app.store.CreateUser(context.Background(), "yash", passwordHash, "member")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repositoryMeta, err := repositories.CreateRepository(context.Background(), store.CreateRepositoryParams{
		Owner:         "yash",
		Name:          "forge",
		Visibility:    "private",
		DefaultBranch: "main",
	})
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	clientPublic, clientPrivate := generateRSAKey(t)
	if _, err := app.store.CreateSSHKey(context.Background(), store.CreateSSHKeyParams{
		UserID:            user.ID,
		Name:              "test-key",
		PublicKey:         strings.TrimSpace(string(cryptossh.MarshalAuthorizedKey(clientPublic))),
		FingerprintSHA256: cryptossh.FingerprintSHA256(clientPublic),
	}); err != nil {
		t.Fatalf("create ssh key: %v", err)
	}

	seedRepository(t, repositoryMeta.RepoPath)

	cfg := app.cfg
	cfg.SSHHostKeyPath = filepath.Join(t.TempDir(), "host_ed25519")
	cfg.SSHUser = "git"
	sshServer, err := sshgit.New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), app.store, repositories)
	if err != nil {
		t.Fatalf("new ssh server: %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen ssh: %v", err)
	}
	defer listener.Close()

	go func() {
		_ = sshServer.Serve(t.Context(), listener)
	}()

	privateKeyPath := filepath.Join(t.TempDir(), "client_key")
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(clientPrivate)
	if err := os.WriteFile(privateKeyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}), 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}

	remote := "ssh://git@" + listener.Addr().String() + "/yash/forge.git"
	sshCommand := "ssh -i " + privateKeyPath + " -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p " + strings.Split(listener.Addr().String(), ":")[1]
	output := runGit(t, "", []string{"GIT_SSH_COMMAND=" + sshCommand}, "ls-remote", remote)
	if !strings.Contains(output, "refs/heads/main") {
		t.Fatalf("expected ls-remote to include main branch, output = %s", output)
	}
}

func TestGitSmartHTTPOrgMaintainerPushAndLsRemote(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	app, repositories := newTestServer(t)
	aliceHash, err := auth.HashPassword("alice-password")
	if err != nil {
		t.Fatalf("hash alice password: %v", err)
	}
	alice, err := app.store.CreateUser(context.Background(), "alice", aliceHash, "member")
	if err != nil {
		t.Fatalf("create alice: %v", err)
	}
	bobHash, err := auth.HashPassword("bob-password")
	if err != nil {
		t.Fatalf("hash bob password: %v", err)
	}
	if _, err := app.store.CreateUser(context.Background(), "bob", bobHash, "member"); err != nil {
		t.Fatalf("create bob: %v", err)
	}

	if _, err := app.store.CreateOrganization(context.Background(), store.CreateOrganizationParams{
		Slug:        "team",
		DisplayName: "Team",
		Description: "shared",
		CreatedBy:   alice.ID,
	}); err != nil {
		t.Fatalf("create org: %v", err)
	}
	if _, err := app.store.AddOrganizationMember(context.Background(), store.AddOrganizationMemberParams{
		OrganizationSlug: "team",
		Username:         "bob",
		Role:             store.OrganizationRoleMaintainer,
	}); err != nil {
		t.Fatalf("add bob to org: %v", err)
	}
	if _, err := repositories.CreateRepository(context.Background(), store.CreateRepositoryParams{
		Owner:         "team",
		OwnerType:     store.OwnerTypeOrganization,
		Name:          "forge",
		Visibility:    "private",
		DefaultBranch: "main",
	}); err != nil {
		t.Fatalf("create org repository: %v", err)
	}

	server := httptest.NewServer(app.Router())
	defer server.Close()

	worktree := t.TempDir()
	runGit(t, worktree, nil, "init")
	runGit(t, worktree, nil, "config", "user.email", "bob@example.com")
	runGit(t, worktree, nil, "config", "user.name", "Bob")
	if err := os.WriteFile(filepath.Join(worktree, "README.md"), []byte("forge\n"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	runGit(t, worktree, nil, "add", "README.md")
	runGit(t, worktree, nil, "commit", "-m", "initial")

	remote := strings.Replace(server.URL, "http://", "http://bob:bob-password@", 1) + "/git/team/forge.git"
	runGit(t, worktree, []string{"GIT_TERMINAL_PROMPT=0"}, "remote", "add", "origin", remote)
	runGit(t, worktree, []string{"GIT_TERMINAL_PROMPT=0"}, "push", "origin", "HEAD:refs/heads/main")

	output := runGit(t, "", []string{"GIT_TERMINAL_PROMPT=0"}, "ls-remote", remote)
	if !strings.Contains(output, "refs/heads/main") {
		t.Fatalf("expected ls-remote to include main branch, output = %s", output)
	}
}

func TestGitSSHCollaboratorReadAccess(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	if _, err := exec.LookPath("ssh"); err != nil {
		t.Skip("ssh not available")
	}

	app, repositories := newTestServer(t)
	aliceHash, err := auth.HashPassword("alice-password")
	if err != nil {
		t.Fatalf("hash alice password: %v", err)
	}
	if _, err := app.store.CreateUser(context.Background(), "alice", aliceHash, "member"); err != nil {
		t.Fatalf("create alice: %v", err)
	}
	bobHash, err := auth.HashPassword("bob-password")
	if err != nil {
		t.Fatalf("hash bob password: %v", err)
	}
	bob, err := app.store.CreateUser(context.Background(), "bob", bobHash, "member")
	if err != nil {
		t.Fatalf("create bob: %v", err)
	}

	repositoryMeta, err := repositories.CreateRepository(context.Background(), store.CreateRepositoryParams{
		Owner:         "alice",
		OwnerType:     store.OwnerTypeUser,
		Name:          "shared",
		Visibility:    "private",
		DefaultBranch: "main",
	})
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}
	if _, err := app.store.AddRepositoryCollaborator(context.Background(), store.AddRepositoryCollaboratorParams{
		Owner:    "alice",
		RepoName: "shared",
		Username: "bob",
		Role:     store.RepositoryRoleRead,
	}); err != nil {
		t.Fatalf("add collaborator: %v", err)
	}

	clientPublic, clientPrivate := generateRSAKey(t)
	if _, err := app.store.CreateSSHKey(context.Background(), store.CreateSSHKeyParams{
		UserID:            bob.ID,
		Name:              "test-key",
		PublicKey:         strings.TrimSpace(string(cryptossh.MarshalAuthorizedKey(clientPublic))),
		FingerprintSHA256: cryptossh.FingerprintSHA256(clientPublic),
	}); err != nil {
		t.Fatalf("create ssh key: %v", err)
	}

	seedRepository(t, repositoryMeta.RepoPath)

	cfg := app.cfg
	cfg.SSHHostKeyPath = filepath.Join(t.TempDir(), "host_ed25519")
	cfg.SSHUser = "git"
	sshServer, err := sshgit.New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), app.store, repositories)
	if err != nil {
		t.Fatalf("new ssh server: %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen ssh: %v", err)
	}
	defer listener.Close()

	go func() {
		_ = sshServer.Serve(t.Context(), listener)
	}()

	privateKeyPath := filepath.Join(t.TempDir(), "client_key")
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(clientPrivate)
	if err := os.WriteFile(privateKeyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}), 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}

	remote := "ssh://git@" + listener.Addr().String() + "/alice/shared.git"
	sshCommand := "ssh -i " + privateKeyPath + " -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p " + strings.Split(listener.Addr().String(), ":")[1]
	output := runGit(t, "", []string{"GIT_SSH_COMMAND=" + sshCommand}, "ls-remote", remote)
	if !strings.Contains(output, "refs/heads/main") {
		t.Fatalf("expected ls-remote to include main branch, output = %s", output)
	}
}

func newTestServer(t *testing.T) (*Server, *repository.Service) {
	t.Helper()

	reposRoot := t.TempDir()
	cfg := config.Config{
		Environment:         "test",
		BaseURL:             "http://localhost:3000",
		CookieName:          "forge_session",
		ReposRoot:           reposRoot,
		Secret:              "test-secret",
		SessionTTL:          time.Hour,
		ReadTimeout:         time.Second,
		WriteTimeout:        time.Second,
		IdleTimeout:         time.Second,
		ShutdownTimeout:     time.Second,
		RequestTimeout:      time.Second,
		MaxRequestBodyBytes: 1 << 20,
		SSHEnabled:          true,
		SSHAddress:          ":2222",
		SSHUser:             "git",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	st := memory.NewStore()
	repositories, err := repository.NewService(logger, st, reposRoot)
	if err != nil {
		t.Fatalf("new repository service: %v", err)
	}
	repositories.Start(t.Context())

	app, err := New(cfg, logger, st, repositories)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	return app, repositories
}

func runGit(t *testing.T, workdir string, extraEnv []string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	if workdir != "" {
		cmd.Dir = workdir
	}
	cmd.Env = append(os.Environ(), extraEnv...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
	return string(output)
}

func generateRSAKey(t *testing.T) (cryptossh.PublicKey, *rsa.PrivateKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	sshPublicKey, err := cryptossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("new ssh public key: %v", err)
	}
	return sshPublicKey, privateKey
}

func seedRepository(t *testing.T, repoPath string) {
	t.Helper()

	worktree := t.TempDir()
	runGit(t, worktree, nil, "init")
	runGit(t, worktree, nil, "config", "user.email", "yash@example.com")
	runGit(t, worktree, nil, "config", "user.name", "Yash")
	if err := os.WriteFile(filepath.Join(worktree, "README.md"), []byte("forge\n"), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	runGit(t, worktree, nil, "add", "README.md")
	runGit(t, worktree, nil, "commit", "-m", "seed")
	runGit(t, worktree, nil, "push", repoPath, "HEAD:refs/heads/main")
}
