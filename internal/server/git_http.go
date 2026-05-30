package server

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yashlunawat/forge/internal/store"
)

const gitBasicRealm = `Basic realm="Forge Git"`

func (s *Server) handleGitHTTP(w http.ResponseWriter, r *http.Request) {
	owner, repoName, suffix, err := parseGitPath(r.URL.Path)
	if err != nil {
		s.writeError(r, w, http.StatusNotFound, err)
		return
	}

	repository, err := s.repositories.GetRepository(r.Context(), owner, repoName)
	if err != nil {
		s.writeError(r, w, http.StatusNotFound, errors.New("repository not found"))
		return
	}

	currentUser, _ := s.authenticateRequest(r)
	writeAccess := isGitWriteRequest(r)
	if writeAccess {
		if !s.repositories.CanWrite(currentUser, repository) {
			w.Header().Set("WWW-Authenticate", gitBasicRealm)
			s.writeError(r, w, http.StatusUnauthorized, errors.New("git write requires valid credentials"))
			return
		}
	} else if !s.repositories.CanRead(currentUser, repository) {
		w.Header().Set("WWW-Authenticate", gitBasicRealm)
		s.writeError(r, w, http.StatusUnauthorized, errors.New("git read requires valid credentials"))
		return
	}

	relativeRepoPath, err := s.repositories.RelativeRepoPath(repository)
	if err != nil {
		s.writeError(r, w, http.StatusInternalServerError, errInternal)
		return
	}
	if err := s.serveGitBackend(w, r, relativeRepoPath, suffix, currentUser); err != nil {
		s.logger.Error("git http backend", "error", err, "owner", repository.Owner, "repo", repository.Name, "suffix", suffix)
		s.writeError(r, w, http.StatusBadGateway, errors.New("git backend failed"))
		return
	}

	if writeAccess && strings.HasSuffix(suffix, "/git-receive-pack") {
		s.repositories.ScheduleMaintenance(repository)
	}
}

func (s *Server) serveGitBackend(w http.ResponseWriter, r *http.Request, relativeRepoPath, suffix string, currentUser *store.User) error {
	cmd := s.newGitHTTPBackendCommand(r.Context())
	cmd.Dir = s.cfg.ReposRoot
	cmd.Stdin = r.Body

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	cmd.Env = append(os.Environ(),
		"GIT_PROJECT_ROOT="+s.cfg.ReposRoot,
		"GIT_HTTP_EXPORT_ALL=1",
		"REQUEST_METHOD="+r.Method,
		"QUERY_STRING="+r.URL.RawQuery,
		"PATH_INFO="+path.Join("/", relativeRepoPath, suffix),
		"CONTENT_TYPE="+r.Header.Get("Content-Type"),
		"REMOTE_ADDR="+r.RemoteAddr,
		"CONTENT_LENGTH="+contentLengthHeader(r),
	)
	if currentUser != nil {
		cmd.Env = append(cmd.Env,
			"REMOTE_USER="+currentUser.Username,
			"AUTH_TYPE=Basic",
		)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	reader := bufio.NewReader(stdout)
	headers, err := readCGIHeaders(reader)
	if err != nil {
		_ = cmd.Wait()
		if stderr.Len() > 0 {
			s.logger.Warn("git-http-backend stderr", "stderr", strings.TrimSpace(stderr.String()))
		}
		return err
	}

	statusCode := http.StatusOK
	for key, values := range headers {
		if strings.EqualFold(key, "Status") {
			statusCode = parseStatus(values[0])
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.WriteHeader(statusCode)

	if _, err := io.Copy(w, reader); err != nil {
		_ = cmd.Wait()
		return err
	}
	if err := cmd.Wait(); err != nil {
		if stderr.Len() > 0 {
			s.logger.Warn("git-http-backend stderr", "stderr", strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
}

func (s *Server) newGitHTTPBackendCommand(ctx context.Context) *exec.Cmd {
	if backend := findGitHTTPBackendBinary(); backend != "" {
		return exec.CommandContext(ctx, backend)
	}
	return exec.CommandContext(ctx, "git", "http-backend")
}

func parseGitPath(requestPath string) (owner, repoName, suffix string, err error) {
	trimmed := strings.TrimPrefix(requestPath, "/git/")
	segments := strings.Split(trimmed, "/")
	if len(segments) < 2 {
		return "", "", "", errors.New("invalid git path")
	}

	owner = segments[0]
	repoSegment := segments[1]
	if !strings.HasSuffix(repoSegment, ".git") {
		return "", "", "", errors.New("invalid repository path")
	}
	repoName = strings.TrimSuffix(repoSegment, ".git")
	if len(segments) > 2 {
		suffix = "/" + strings.Join(segments[2:], "/")
	}
	return owner, repoName, suffix, nil
}

func isGitWriteRequest(r *http.Request) bool {
	service := r.URL.Query().Get("service")
	return service == "git-receive-pack" || strings.HasSuffix(r.URL.Path, "/git-receive-pack")
}

func readCGIHeaders(reader *bufio.Reader) (textproto.MIMEHeader, error) {
	tp := textproto.NewReader(reader)
	return tp.ReadMIMEHeader()
}

func parseStatus(value string) int {
	parts := strings.Fields(strings.TrimSpace(value))
	if len(parts) == 0 {
		return http.StatusOK
	}
	if code, err := strconv.Atoi(parts[0]); err == nil {
		return code
	}
	return http.StatusOK
}

func contentLengthHeader(r *http.Request) string {
	if r.ContentLength <= 0 {
		return ""
	}
	return strconv.FormatInt(r.ContentLength, 10)
}

func findGitHTTPBackendBinary() string {
	candidates := []string{
		os.Getenv("FORGE_GIT_HTTP_BACKEND_BIN"),
		"git-http-backend",
		"/usr/libexec/git-core/git-http-backend",
		"/usr/lib/git-core/git-http-backend",
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if filepath.IsAbs(candidate) {
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
			continue
		}
		if resolved, err := exec.LookPath(candidate); err == nil {
			return resolved
		}
	}
	return ""
}
