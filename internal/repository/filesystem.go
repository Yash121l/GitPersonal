package repository

import (
	"context"
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
	return filepath.Join(p.root, slug(owner), slug(name)+".git")
}

func (p *FilesystemProvisioner) CreateBareRepository(ctx context.Context, owner, name, defaultBranch string) (string, error) {
	repoPath := p.RepoPath(owner, name)
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	if err := os.MkdirAll(filepath.Dir(repoPath), 0o755); err != nil {
		return "", fmt.Errorf("create repo parent directory: %w", err)
	}
	if _, err := os.Stat(repoPath); err == nil {
		return "", os.ErrExist
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat repo path: %w", err)
	}

	command := exec.CommandContext(ctx, p.gitBin, "init", "--bare", "--quiet", repoPath)
	if output, err := command.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git init --bare: %w: %s", err, strings.TrimSpace(string(output)))
	}

	headPath := filepath.Join(repoPath, "HEAD")
	headContents := []byte("ref: refs/heads/" + defaultBranch + "\n")
	if err := os.WriteFile(headPath, headContents, 0o644); err != nil {
		_ = os.RemoveAll(repoPath)
		return "", fmt.Errorf("write HEAD file: %w", err)
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
