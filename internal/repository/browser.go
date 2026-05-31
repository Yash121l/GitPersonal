package repository

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/yashlunawat/forge/internal/store"
)

const maxBlobPreviewBytes int64 = 256 * 1024

type Branch struct {
	Name string `json:"name"`
}

type TreeEntry struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Mode     string `json:"mode"`
	ObjectID string `json:"object_id"`
	SizeBytes int64 `json:"size_bytes"`
}

type Blob struct {
	Path       string `json:"path"`
	SizeBytes  int64  `json:"size_bytes"`
	Content    string `json:"content"`
	Truncated  bool   `json:"truncated"`
	IsBinary   bool   `json:"is_binary"`
	Language   string `json:"language"`
}

func (s *Service) ListBranches(ctx context.Context, repository store.Repository) ([]Branch, error) {
	output, err := s.provisioner.RunGit(ctx, repository.RepoPath, "for-each-ref", "--format=%(refname:short)", "refs/heads")
	if err != nil {
		return nil, fmt.Errorf("list branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	branches := make([]Branch, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		branches = append(branches, Branch{Name: line})
	}
	return branches, nil
}

func (s *Service) ListTree(ctx context.Context, repository store.Repository, ref, treePath string) ([]TreeEntry, error) {
	ref = normalizeBrowserRef(ref, repository.DefaultBranch)
	normalizedPath, err := normalizeBrowserPath(treePath)
	if err != nil {
		return nil, err
	}

	target := ref
	if normalizedPath != "" {
		target = ref + ":" + normalizedPath
	}

	output, err := s.provisioner.RunGit(ctx, repository.RepoPath, "ls-tree", "-z", "-l", target)
	if err != nil {
		return nil, mapGitBrowserError("list tree", err)
	}

	chunks := bytes.Split([]byte(output), []byte{0})
	entries := make([]TreeEntry, 0, len(chunks))
	for _, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}

		tabIndex := bytes.IndexByte(chunk, '\t')
		if tabIndex < 0 {
			return nil, errors.New("parse tree entry: missing tab separator")
		}

		header := strings.Fields(string(chunk[:tabIndex]))
		if len(header) < 4 {
			return nil, errors.New("parse tree entry: invalid header")
		}

		sizeBytes := int64(0)
		if len(header) >= 4 && header[3] != "-" {
			sizeBytes, err = strconv.ParseInt(header[3], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse tree entry size: %w", err)
			}
		}

		name := string(chunk[tabIndex+1:])
		entryPath := name
		if normalizedPath != "" {
			entryPath = path.Join(normalizedPath, name)
		}

		entries = append(entries, TreeEntry{
			Path:      entryPath,
			Name:      name,
			Type:      header[1],
			Mode:      header[0],
			ObjectID:  header[2],
			SizeBytes: sizeBytes,
		})
	}

	return entries, nil
}

func (s *Service) ReadBlob(ctx context.Context, repository store.Repository, ref, blobPath string) (Blob, error) {
	ref = normalizeBrowserRef(ref, repository.DefaultBranch)
	normalizedPath, err := normalizeBrowserPath(blobPath)
	if err != nil {
		return Blob{}, err
	}
	if normalizedPath == "" {
		return Blob{}, store.ErrInvalidArgument
	}

	objectSpec := ref + ":" + normalizedPath
	sizeOutput, err := s.provisioner.RunGit(ctx, repository.RepoPath, "cat-file", "-s", objectSpec)
	if err != nil {
		return Blob{}, mapGitBrowserError("read blob size", err)
	}

	sizeBytes, err := strconv.ParseInt(strings.TrimSpace(sizeOutput), 10, 64)
	if err != nil {
		return Blob{}, fmt.Errorf("parse blob size: %w", err)
	}

	content, truncated, stderr, err := s.provisioner.RunGitLimited(ctx, repository.RepoPath, maxBlobPreviewBytes, "show", objectSpec)
	if err != nil {
		return Blob{}, mapGitBrowserErrorWithDetails("read blob", err, stderr)
	}

	isBinary := bytes.IndexByte(content, 0) >= 0 || !utf8.Valid(content)
	if isBinary {
		return Blob{
			Path:      normalizedPath,
			SizeBytes: sizeBytes,
			Truncated: truncated,
			IsBinary:  true,
			Language:  detectLanguage(normalizedPath),
		}, nil
	}

	return Blob{
		Path:      normalizedPath,
		SizeBytes: sizeBytes,
		Content:   string(content),
		Truncated: truncated || sizeBytes > maxBlobPreviewBytes,
		IsBinary:  false,
		Language:  detectLanguage(normalizedPath),
	}, nil
}

func normalizeBrowserRef(ref, fallback string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		ref = strings.TrimSpace(fallback)
	}
	if ref == "" {
		return "HEAD"
	}
	return ref
}

func normalizeBrowserPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	trimmed := strings.TrimPrefix(value, "/")
	for _, segment := range strings.Split(trimmed, "/") {
		if segment == "." || segment == ".." {
			return "", store.ErrInvalidArgument
		}
	}

	cleaned := path.Clean("/" + trimmed)
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "." {
		return "", nil
	}
	for _, segment := range strings.Split(cleaned, "/") {
		if segment == "." || segment == ".." || segment == "" {
			return "", store.ErrInvalidArgument
		}
	}
	return cleaned, nil
}

func mapGitBrowserError(operation string, err error) error {
	return mapGitBrowserErrorWithDetails(operation, err, "")
}

func mapGitBrowserErrorWithDetails(operation string, err error, details string) error {
	if err == nil {
		return nil
	}
	details = strings.ToLower(details)
	if strings.Contains(details, "not a valid object name") || strings.Contains(details, "path does not exist") || strings.Contains(details, "exists on disk, but not in") || strings.Contains(details, "fatal: invalid object name") {
		return store.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func detectLanguage(filePath string) string {
	switch {
	case strings.HasSuffix(filePath, ".go"):
		return "go"
	case strings.HasSuffix(filePath, ".ts"):
		return "typescript"
	case strings.HasSuffix(filePath, ".tsx"):
		return "tsx"
	case strings.HasSuffix(filePath, ".js"):
		return "javascript"
	case strings.HasSuffix(filePath, ".vue"):
		return "vue"
	case strings.HasSuffix(filePath, ".json"):
		return "json"
	case strings.HasSuffix(filePath, ".md"):
		return "markdown"
	case strings.HasSuffix(filePath, ".yml"), strings.HasSuffix(filePath, ".yaml"):
		return "yaml"
	case strings.HasSuffix(filePath, ".html"):
		return "html"
	case strings.HasSuffix(filePath, ".css"):
		return "css"
	case strings.HasSuffix(filePath, ".sh"):
		return "shell"
	default:
		return "text"
	}
}
