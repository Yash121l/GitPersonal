package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type FilesystemProvisioner struct {
	root   string
	gitBin string
}

func NewFilesystemProvisioner(root string) *FilesystemProvisioner {
	return &FilesystemProvisioner{
		root:   root,
		gitBin: "git",
	}
}

func (p *FilesystemProvisioner) RepoPath(owner, name string) string {
	key := slug(owner) + "/" + slug(name)
	sum := sha256.Sum256([]byte(key))
	prefixA := hex.EncodeToString(sum[:1])
	prefixB := hex.EncodeToString(sum[1:2])

	return filepath.Join(p.root, prefixA, prefixB, slug(owner), slug(name)+".git")
}

func (p *FilesystemProvisioner) CreateBareRepository(ctx context.Context, owner, name, defaultBranch string) (string, error) {
	if err := p.EnsureStorageLayout(); err != nil {
		return "", err
	}

	repoPath := p.RepoPath(owner, name)
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	if _, err := os.Stat(repoPath); err == nil {
		return "", os.ErrExist
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat repo path: %w", err)
	}

	stagingPath := filepath.Join(p.root, ".staging", fmt.Sprintf("%d-%s.git", time.Now().UTC().UnixNano(), slug(name)))
	if output, err := p.runGit(ctx, "init", "--bare", "--quiet", "--shared=group", stagingPath); err != nil {
		return "", fmt.Errorf("git init --bare: %w: %s", err, strings.TrimSpace(output))
	}
	if err := p.configureRepository(ctx, stagingPath); err != nil {
		_ = os.RemoveAll(stagingPath)
		return "", err
	}

	headPath := filepath.Join(stagingPath, "HEAD")
	headContents := []byte("ref: refs/heads/" + defaultBranch + "\n")
	if err := os.WriteFile(headPath, headContents, 0o644); err != nil {
		_ = os.RemoveAll(stagingPath)
		return "", fmt.Errorf("write HEAD file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(repoPath), 0o755); err != nil {
		_ = os.RemoveAll(stagingPath)
		return "", fmt.Errorf("create repo parent directory: %w", err)
	}
	if err := os.Rename(stagingPath, repoPath); err != nil {
		_ = os.RemoveAll(stagingPath)
		if errors.Is(err, os.ErrExist) {
			return "", os.ErrExist
		}
		return "", fmt.Errorf("move repo into place: %w", err)
	}

	return repoPath, nil
}

func (p *FilesystemProvisioner) CleanupRepository(path string) error {
	if path == "" {
		return nil
	}
	if err := p.ensureManagedPath(path); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func (p *FilesystemProvisioner) StageDelete(path string) (string, func() error, error) {
	if path == "" {
		return "", func() error { return nil }, nil
	}
	if err := p.ensureManagedPath(path); err != nil {
		return "", nil, err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "", func() error { return nil }, nil
	} else if err != nil {
		return "", nil, fmt.Errorf("stat repo path: %w", err)
	}

	trashDir := filepath.Join(p.root, ".trash")
	if err := os.MkdirAll(trashDir, 0o755); err != nil {
		return "", nil, fmt.Errorf("create trash dir: %w", err)
	}

	trashPath := filepath.Join(trashDir, fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), filepath.Base(path)))
	if err := os.Rename(path, trashPath); err != nil {
		return "", nil, fmt.Errorf("move repo to trash: %w", err)
	}

	restore := func() error {
		if _, err := os.Stat(trashPath); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return os.Rename(trashPath, path)
	}

	return trashPath, restore, nil
}

func (p *FilesystemProvisioner) EnsureStorageLayout() error {
	for _, path := range []string{
		p.root,
		filepath.Join(p.root, ".staging"),
		filepath.Join(p.root, ".trash"),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create storage path %s: %w", path, err)
		}
	}

	return nil
}

func (p *FilesystemProvisioner) Check(ctx context.Context) error {
	if _, err := exec.LookPath(p.gitBin); err != nil {
		return fmt.Errorf("resolve git binary: %w", err)
	}
	if err := p.EnsureStorageLayout(); err != nil {
		return err
	}

	probeDir, err := os.MkdirTemp(filepath.Join(p.root, ".staging"), "probe-*")
	if err != nil {
		return fmt.Errorf("write probe in staging area: %w", err)
	}
	return os.RemoveAll(probeDir)
}

func slug(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (p *FilesystemProvisioner) ensureManagedPath(path string) error {
	root := filepath.Clean(p.root)
	candidate := filepath.Clean(path)

	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return fmt.Errorf("resolve relative path: %w", err)
	}
	if relative == "." {
		return errors.New("refusing to operate on repository root")
	}
	if strings.HasPrefix(relative, "..") {
		return errors.New("path is outside repository root")
	}

	return nil
}

func (p *FilesystemProvisioner) configureRepository(ctx context.Context, repoPath string) error {
	configPairs := [][2]string{
		{"core.sharedRepository", "group"},
		{"gc.auto", "0"},
		{"receive.autogc", "false"},
		{"core.logAllRefUpdates", "true"},
		{"repack.writeBitmaps", "true"},
		{"fetch.writeCommitGraph", "true"},
		{"commitGraph.generationVersion", "2"},
		{"pack.useSparse", "true"},
	}

	for _, pair := range configPairs {
		if output, err := p.runGit(ctx, "--git-dir", repoPath, "config", pair[0], pair[1]); err != nil {
			return fmt.Errorf("configure repo %s=%s: %w: %s", pair[0], pair[1], err, strings.TrimSpace(output))
		}
	}

	return nil
}

func (p *FilesystemProvisioner) runGit(ctx context.Context, args ...string) (string, error) {
	command := exec.CommandContext(ctx, p.gitBin, args...)
	output, err := command.CombinedOutput()
	return string(output), err
}
