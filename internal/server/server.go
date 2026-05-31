package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	cryptossh "golang.org/x/crypto/ssh"

	"github.com/yashlunawat/forge/internal/auth"
	"github.com/yashlunawat/forge/internal/config"
	"github.com/yashlunawat/forge/internal/repository"
	"github.com/yashlunawat/forge/internal/store"
)

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

type Server struct {
	cfg          config.Config
	logger       *slog.Logger
	store        store.Store
	repositories *repository.Service
	router       http.Handler
}

type contextKey string

const userContextKey contextKey = "user"

type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type createRepoRequest struct {
	Owner         string `json:"owner"`
	OwnerType     string `json:"owner_type"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Visibility    string `json:"visibility"`
	DefaultBranch string `json:"default_branch"`
}

type createOrganizationRequest struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

type addOrganizationMemberRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type createSSHKeyRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type addRepositoryCollaboratorRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type createRepositoryWebhookRequest struct {
	URL    string   `json:"url"`
	Secret string   `json:"secret"`
	Events []string `json:"events"`
}

type currentUserResponse struct {
	User store.User `json:"user"`
}

type repoListResponse struct {
	Repositories []store.Repository `json:"repositories"`
}

type organizationListResponse struct {
	Organizations []store.OrganizationMembership `json:"organizations"`
}

type sshKeyListResponse struct {
	Keys []store.SSHKey `json:"keys"`
}

type repositoryDetailResponse struct {
	Repository   store.Repository `json:"repository"`
	HTTPCloneURL string           `json:"http_clone_url"`
	SSHCloneURL  string           `json:"ssh_clone_url,omitempty"`
}

type repositoryBranchListResponse struct {
	Branches []repository.Branch `json:"branches"`
}

type repositoryTreeResponse struct {
	Ref     string                 `json:"ref"`
	Path    string                 `json:"path"`
	Entries []repository.TreeEntry `json:"entries"`
}

type repositoryBlobResponse struct {
	Ref  string          `json:"ref"`
	Blob repository.Blob `json:"blob"`
}

type repositoryWebhookListResponse struct {
	Webhooks []store.RepositoryWebhook `json:"webhooks"`
}

func New(cfg config.Config, logger *slog.Logger, st store.Store, repositories *repository.Service) (*Server, error) {
	s := &Server{
		cfg:          cfg,
		logger:       logger,
		store:        st,
		repositories: repositories,
	}
	s.router = s.routes()
	return s, nil
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(s.requestID)
	r.Use(s.requestLogging)
	r.Use(s.recoverer)
	r.Use(s.securityHeaders)
	r.Use(s.enforceBodyLimit)
	r.Use(s.enforceRequestTimeout)

	r.Get("/healthz", s.handleHealthz)
	r.Get("/readyz", s.handleReadyz)
	r.Get("/", s.handleAppEntry)
	r.Get("/app", s.handleAppEntry)
	r.Handle("/app/assets/*", s.handleUIAssets())
	r.Get("/app/login", s.handleUIPage("Forge | Sign In").ServeHTTP)
	r.Get("/app/register", s.handleUIPage("Forge | Create Account").ServeHTTP)
	r.With(s.requireAppSession).Get("/app/*", s.handleUIPage("Forge").ServeHTTP)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", s.handleRegister)
			r.Post("/login", s.handleLogin)
			r.Post("/logout", s.handleLogout)
		})

		r.With(s.requireAuth).Get("/me", s.handleCurrentUser)
		r.With(s.requireAuth).Get("/keys", s.handleListSSHKeys)
		r.With(s.requireAuth).Post("/keys", s.handleCreateSSHKey)

		r.With(s.requireAuth).Get("/orgs", s.handleListOrganizations)
		r.With(s.requireAuth).Post("/orgs", s.handleCreateOrganization)
		r.With(s.requireAuth).Post("/orgs/{org}/members", s.handleAddOrganizationMember)

		r.With(s.requireAuth).Get("/repos", s.handleListRepositories)
		r.With(s.requireAuth).Get("/repos/{owner}/{repo}", s.handleGetRepository)
		r.With(s.requireAuth).Get("/repos/{owner}/{repo}/branches", s.handleListRepositoryBranches)
		r.With(s.requireAuth).Get("/repos/{owner}/{repo}/tree", s.handleGetRepositoryTree)
		r.With(s.requireAuth).Get("/repos/{owner}/{repo}/blob", s.handleGetRepositoryBlob)
		r.With(s.requireAuth).Post("/repos", s.handleCreateRepository)
		r.With(s.requireAuth).Delete("/repos/{owner}/{repo}", s.handleDeleteRepository)
		r.With(s.requireAuth).Post("/repos/{owner}/{repo}/collaborators", s.handleAddRepositoryCollaborator)
		r.With(s.requireAuth).Get("/repos/{owner}/{repo}/webhooks", s.handleListRepositoryWebhooks)
		r.With(s.requireAuth).Post("/repos/{owner}/{repo}/webhooks", s.handleCreateRepositoryWebhook)
		r.With(s.requireAuth).Delete("/repos/{owner}/{repo}/webhooks/{webhookID}", s.handleDeleteRepositoryWebhook)
	})

	r.Handle("/git/*", http.HandlerFunc(s.handleGitHTTP))

	return r
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"name":   "forge",
	})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	checkCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.store.Check(checkCtx); err != nil {
		s.writeError(r, w, http.StatusServiceUnavailable, errors.New("database not ready"))
		return
	}
	if err := s.repositories.Check(checkCtx); err != nil {
		s.writeError(r, w, http.StatusServiceUnavailable, errors.New("repository storage not ready"))
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
		"name":   "forge",
	})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if err := validateUsername(req.Username); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}
	if len(req.Password) < 12 {
		s.writeError(r, w, http.StatusBadRequest, errors.New("password must be at least 12 characters"))
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("hash password", "error", err)
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to create user"))
		return
	}

	user, err := s.store.CreateUser(r.Context(), req.Username, passwordHash, "member")
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrAlreadyExists) {
			status = http.StatusConflict
		}
		s.writeError(r, w, status, err)
		return
	}

	if err := s.setSessionCookie(r.Context(), w, user); err != nil {
		s.logger.Error("set session cookie", "error", err)
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to start session"))
		return
	}

	s.writeJSON(w, http.StatusCreated, currentUserResponse{User: user})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}

	user, err := s.store.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("invalid credentials"))
		return
	}
	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("invalid credentials"))
		return
	}

	if err := s.setSessionCookie(r.Context(), w, user); err != nil {
		s.logger.Error("set session cookie", "error", err)
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to start session"))
		return
	}

	s.writeJSON(w, http.StatusOK, currentUserResponse{User: user})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(s.cfg.CookieName); err == nil {
		if claims, err := auth.ParseToken(s.cfg.Secret, cookie.Value); err == nil && claims.ID != "" {
			if err := s.store.RevokeSession(r.Context(), claims.ID, time.Now().UTC()); err != nil && !errors.Is(err, store.ErrNotFound) {
				s.logger.Warn("revoke session", "error", err, "token_id", claims.ID)
			}
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		Secure:   strings.HasPrefix(s.cfg.BaseURL, "https://"),
	})

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (s *Server) handleCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	s.writeJSON(w, http.StatusOK, currentUserResponse{User: user})
}

func (s *Server) handleListRepositories(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	var (
		repositories []store.Repository
		err          error
	)
	if user.Role == store.OrganizationRoleOwner {
		repositories, err = s.store.ListRepositories(r.Context())
	} else {
		repositories, err = s.store.ListRepositoriesForUser(r.Context(), user.ID)
	}
	if err != nil {
		s.logger.Error("list repositories", "error", err, "user_id", user.ID)
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to list repositories"))
		return
	}
	if repositories == nil {
		repositories = []store.Repository{}
	}

	s.writeJSON(w, http.StatusOK, repoListResponse{Repositories: repositories})
}

func (s *Server) handleGetRepository(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "repo")
	repository, err := s.repositories.GetRepository(r.Context(), owner, repoName)
	if err != nil {
		s.writeError(r, w, http.StatusNotFound, errors.New("repository not found"))
		return
	}

	canRead, err := s.repositories.CanRead(r.Context(), &user, repository)
	if err != nil {
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to authorize repository access"))
		return
	}
	if !canRead {
		s.writeError(r, w, http.StatusForbidden, errors.New("repository read access required"))
		return
	}

	httpCloneURL, sshCloneURL := s.cloneURLs(repository.Owner, repository.Name)
	s.writeJSON(w, http.StatusOK, repositoryDetailResponse{
		Repository:   repository,
		HTTPCloneURL: httpCloneURL,
		SSHCloneURL:  sshCloneURL,
	})
}

func (s *Server) handleListRepositoryBranches(w http.ResponseWriter, r *http.Request) {
	_, repoMeta, ok := s.authorizeRepositoryRead(w, r)
	if !ok {
		return
	}

	branches, err := s.repositories.ListBranches(r.Context(), repoMeta)
	if err != nil {
		s.logger.Error("list repository branches", "error", err, "owner", repoMeta.Owner, "repo", repoMeta.Name)
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to list repository branches"))
		return
	}
	if branches == nil {
		branches = []repository.Branch{}
	}

	s.writeJSON(w, http.StatusOK, repositoryBranchListResponse{Branches: branches})
}

func (s *Server) handleGetRepositoryTree(w http.ResponseWriter, r *http.Request) {
	_, repoMeta, ok := s.authorizeRepositoryRead(w, r)
	if !ok {
		return
	}

	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	treePath := strings.TrimSpace(r.URL.Query().Get("path"))
	entries, err := s.repositories.ListTree(r.Context(), repoMeta, ref, treePath)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, store.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, store.ErrInvalidArgument):
			status = http.StatusBadRequest
		}
		s.writeError(r, w, status, err)
		return
	}

	if ref == "" {
		ref = repoMeta.DefaultBranch
	}
	if entries == nil {
		entries = []repository.TreeEntry{}
	}
	s.writeJSON(w, http.StatusOK, repositoryTreeResponse{
		Ref:     ref,
		Path:    treePath,
		Entries: entries,
	})
}

func (s *Server) handleGetRepositoryBlob(w http.ResponseWriter, r *http.Request) {
	_, repoMeta, ok := s.authorizeRepositoryRead(w, r)
	if !ok {
		return
	}

	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	blobPath := strings.TrimSpace(r.URL.Query().Get("path"))
	if blobPath == "" {
		s.writeError(r, w, http.StatusBadRequest, errors.New("blob path is required"))
		return
	}

	blob, err := s.repositories.ReadBlob(r.Context(), repoMeta, ref, blobPath)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, store.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, store.ErrInvalidArgument):
			status = http.StatusBadRequest
		}
		s.writeError(r, w, status, err)
		return
	}

	if ref == "" {
		ref = repoMeta.DefaultBranch
	}
	s.writeJSON(w, http.StatusOK, repositoryBlobResponse{Ref: ref, Blob: blob})
}

func (s *Server) handleListOrganizations(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	organizations, err := s.store.ListOrganizationsByMember(r.Context(), user.ID)
	if err != nil {
		s.logger.Error("list organizations", "error", err, "user_id", user.ID)
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to list organizations"))
		return
	}
	if organizations == nil {
		organizations = []store.OrganizationMembership{}
	}

	s.writeJSON(w, http.StatusOK, organizationListResponse{Organizations: organizations})
}

func (s *Server) handleCreateOrganization(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	var req createOrganizationRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}

	req.Slug = strings.TrimSpace(req.Slug)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if err := validateUsername(req.Slug); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}
	if req.DisplayName == "" {
		req.DisplayName = req.Slug
	}

	organization, err := s.store.CreateOrganization(r.Context(), store.CreateOrganizationParams{
		Slug:        req.Slug,
		DisplayName: req.DisplayName,
		Description: req.Description,
		CreatedBy:   user.ID,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrAlreadyExists) {
			status = http.StatusConflict
		}
		s.writeError(r, w, status, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, organization)
}

func (s *Server) handleAddOrganizationMember(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	orgSlug := chi.URLParam(r, "org")
	membership, err := s.store.GetOrganizationMembership(r.Context(), orgSlug, user.ID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to load organization membership"))
		return
	}
	if user.Role != store.OrganizationRoleOwner && membership.Role != store.OrganizationRoleOwner {
		s.writeError(r, w, http.StatusForbidden, errors.New("only organization owners can add members"))
		return
	}

	var req addOrganizationMemberRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}
	req.Username = strings.TrimSpace(req.Username)

	role, err := normalizeOrganizationRole(req.Role)
	if err != nil {
		s.writeError(r, w, http.StatusBadRequest, errors.New("organization role must be member, maintainer, or owner"))
		return
	}

	added, err := s.store.AddOrganizationMember(r.Context(), store.AddOrganizationMemberParams{
		OrganizationSlug: orgSlug,
		Username:         req.Username,
		Role:             role,
	})
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, store.ErrAlreadyExists):
			status = http.StatusConflict
		case errors.Is(err, store.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, store.ErrInvalidArgument):
			status = http.StatusBadRequest
		}
		s.writeError(r, w, status, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, added)
}

func (s *Server) handleCreateRepository(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	var req createRepoRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Owner = strings.TrimSpace(req.Owner)
	req.OwnerType = normalizeOwnerType(req.OwnerType)
	req.Visibility = normalizeVisibility(req.Visibility)
	req.DefaultBranch = strings.TrimSpace(req.DefaultBranch)

	if err := validateRepositoryName(req.Name); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}
	if req.DefaultBranch == "" {
		req.DefaultBranch = "main"
	}
	if req.Visibility != "public" && req.Visibility != "private" {
		s.writeError(r, w, http.StatusBadRequest, errors.New("visibility must be public or private"))
		return
	}
	if req.OwnerType != store.OwnerTypeUser && req.OwnerType != store.OwnerTypeOrganization {
		s.writeError(r, w, http.StatusBadRequest, errors.New("owner_type must be user or organization"))
		return
	}
	if req.OwnerType == store.OwnerTypeUser {
		if req.Owner == "" {
			req.Owner = user.Username
		}
		if !strings.EqualFold(req.Owner, user.Username) && user.Role != store.OrganizationRoleOwner {
			s.writeError(r, w, http.StatusForbidden, errors.New("cannot create repository for another user"))
			return
		}
	}
	if req.OwnerType == store.OwnerTypeOrganization {
		if req.Owner == "" {
			s.writeError(r, w, http.StatusBadRequest, errors.New("organization-owned repositories require an owner slug"))
			return
		}
		membership, err := s.store.GetOrganizationMembership(r.Context(), req.Owner, user.ID)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to load organization membership"))
			return
		}
		if user.Role != store.OrganizationRoleOwner && membership.Role != store.OrganizationRoleOwner && membership.Role != store.OrganizationRoleMaintainer {
			s.writeError(r, w, http.StatusForbidden, errors.New("organization repository creation requires maintainer or owner access"))
			return
		}
	}

	repository, err := s.repositories.CreateRepository(r.Context(), store.CreateRepositoryParams{
		Owner:         req.Owner,
		OwnerType:     req.OwnerType,
		Name:          req.Name,
		Description:   req.Description,
		Visibility:    req.Visibility,
		DefaultBranch: req.DefaultBranch,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrAlreadyExists) {
			status = http.StatusConflict
		}
		s.writeError(r, w, status, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, repository)
}

func (s *Server) handleDeleteRepository(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")
	repository, err := s.repositories.GetRepository(r.Context(), owner, repo)
	if err != nil {
		s.writeError(r, w, http.StatusNotFound, errors.New("repository not found"))
		return
	}
	canAdmin, err := s.repositories.CanAdmin(r.Context(), &user, repository)
	if err != nil {
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to authorize repository access"))
		return
	}
	if !canAdmin {
		s.writeError(r, w, http.StatusForbidden, errors.New("repository admin access required"))
		return
	}

	if err := s.repositories.DeleteRepository(r.Context(), owner, repo, &user); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		}
		s.writeError(r, w, status, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAddRepositoryCollaborator(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "repo")
	repository, err := s.repositories.GetRepository(r.Context(), owner, repoName)
	if err != nil {
		s.writeError(r, w, http.StatusNotFound, errors.New("repository not found"))
		return
	}

	canAdmin, err := s.repositories.CanAdmin(r.Context(), &user, repository)
	if err != nil {
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to authorize repository access"))
		return
	}
	if !canAdmin {
		s.writeError(r, w, http.StatusForbidden, errors.New("repository admin access required"))
		return
	}

	var req addRepositoryCollaboratorRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}
	req.Username = strings.TrimSpace(req.Username)

	role, err := normalizeRepositoryRole(req.Role)
	if err != nil {
		s.writeError(r, w, http.StatusBadRequest, errors.New("repository role must be read, write, or admin"))
		return
	}

	collaborator, err := s.store.AddRepositoryCollaborator(r.Context(), store.AddRepositoryCollaboratorParams{
		Owner:    owner,
		RepoName: repoName,
		Username: req.Username,
		Role:     role,
	})
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, store.ErrAlreadyExists):
			status = http.StatusConflict
		case errors.Is(err, store.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, store.ErrInvalidArgument):
			status = http.StatusBadRequest
		}
		s.writeError(r, w, status, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, collaborator)
}

func (s *Server) handleListRepositoryWebhooks(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "repo")
	repository, err := s.repositories.GetRepository(r.Context(), owner, repoName)
	if err != nil {
		s.writeError(r, w, http.StatusNotFound, errors.New("repository not found"))
		return
	}
	canAdmin, err := s.repositories.CanAdmin(r.Context(), &user, repository)
	if err != nil {
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to authorize repository access"))
		return
	}
	if !canAdmin {
		s.writeError(r, w, http.StatusForbidden, errors.New("repository admin access required"))
		return
	}

	webhooks, err := s.store.ListRepositoryWebhooks(r.Context(), owner, repoName)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		}
		s.writeError(r, w, status, err)
		return
	}
	if webhooks == nil {
		webhooks = []store.RepositoryWebhook{}
	}

	s.writeJSON(w, http.StatusOK, repositoryWebhookListResponse{Webhooks: webhooks})
}

func (s *Server) handleCreateRepositoryWebhook(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "repo")
	repository, err := s.repositories.GetRepository(r.Context(), owner, repoName)
	if err != nil {
		s.writeError(r, w, http.StatusNotFound, errors.New("repository not found"))
		return
	}
	canAdmin, err := s.repositories.CanAdmin(r.Context(), &user, repository)
	if err != nil {
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to authorize repository access"))
		return
	}
	if !canAdmin {
		s.writeError(r, w, http.StatusForbidden, errors.New("repository admin access required"))
		return
	}

	var req createRepositoryWebhookRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}
	req.URL = strings.TrimSpace(req.URL)
	if err := validateWebhookURL(req.URL); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}
	events, err := store.NormalizeRepositoryWebhookEvents(req.Events)
	if err != nil {
		s.writeError(r, w, http.StatusBadRequest, errors.New("webhook events must only include repository.push or repository.deleted"))
		return
	}

	webhook, err := s.store.CreateRepositoryWebhook(r.Context(), store.CreateRepositoryWebhookParams{
		Owner:    owner,
		RepoName: repoName,
		URL:      req.URL,
		Secret:   req.Secret,
		Events:   events,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		}
		s.writeError(r, w, status, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, webhook)
}

func (s *Server) handleDeleteRepositoryWebhook(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "repo")
	repository, err := s.repositories.GetRepository(r.Context(), owner, repoName)
	if err != nil {
		s.writeError(r, w, http.StatusNotFound, errors.New("repository not found"))
		return
	}
	canAdmin, err := s.repositories.CanAdmin(r.Context(), &user, repository)
	if err != nil {
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to authorize repository access"))
		return
	}
	if !canAdmin {
		s.writeError(r, w, http.StatusForbidden, errors.New("repository admin access required"))
		return
	}

	webhookID, err := strconv.ParseInt(chi.URLParam(r, "webhookID"), 10, 64)
	if err != nil || webhookID <= 0 {
		s.writeError(r, w, http.StatusBadRequest, errors.New("invalid webhook id"))
		return
	}

	if err := s.store.DeleteRepositoryWebhook(r.Context(), owner, repoName, webhookID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		}
		s.writeError(r, w, status, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleCreateSSHKey(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	var req createSSHKeyRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.PublicKey) == "" {
		s.writeError(r, w, http.StatusBadRequest, errors.New("name and public_key are required"))
		return
	}

	fingerprint, err := sshFingerprint(req.PublicKey)
	if err != nil {
		s.writeError(r, w, http.StatusBadRequest, err)
		return
	}

	key, err := s.store.CreateSSHKey(r.Context(), store.CreateSSHKeyParams{
		UserID:            user.ID,
		Name:              strings.TrimSpace(req.Name),
		PublicKey:         strings.TrimSpace(req.PublicKey),
		FingerprintSHA256: fingerprint,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrAlreadyExists) {
			status = http.StatusConflict
		}
		s.writeError(r, w, status, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, key)
}

func (s *Server) handleListSSHKeys(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	keys, err := s.store.ListSSHKeysByUser(r.Context(), user.ID)
	if err != nil {
		s.logger.Error("list ssh keys", "error", err, "user_id", user.ID)
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to list ssh keys"))
		return
	}
	if keys == nil {
		keys = []store.SSHKey{}
	}

	s.writeJSON(w, http.StatusOK, sshKeyListResponse{Keys: keys})
}

func validateWebhookURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return errors.New("webhook url must be valid")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("webhook url must use http or https")
	}
	if parsed.Host == "" {
		return errors.New("webhook url must include a host")
	}
	return nil
}

func (s *Server) authorizeRepositoryRead(w http.ResponseWriter, r *http.Request) (store.User, store.Repository, bool) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
		return store.User{}, store.Repository{}, false
	}

	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "repo")
	repositoryMeta, err := s.repositories.GetRepository(r.Context(), owner, repoName)
	if err != nil {
		s.writeError(r, w, http.StatusNotFound, errors.New("repository not found"))
		return store.User{}, store.Repository{}, false
	}

	canRead, err := s.repositories.CanRead(r.Context(), &user, repositoryMeta)
	if err != nil {
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to authorize repository access"))
		return store.User{}, store.Repository{}, false
	}
	if !canRead {
		s.writeError(r, w, http.StatusForbidden, errors.New("repository read access required"))
		return store.User{}, store.Repository{}, false
	}

	return user, repositoryMeta, true
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.authenticateRequest(r)
		if err != nil {
			s.writeError(r, w, http.StatusUnauthorized, errors.New("authentication required"))
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, *user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) setSessionCookie(ctx context.Context, w http.ResponseWriter, user store.User) error {
	now := time.Now().UTC()
	tokenID := newOpaqueID()
	token, expiresAt, err := auth.NewTokenWithID(s.cfg.Secret, tokenID, s.cfg.SessionTTL, user.ID, user.Username, user.Role, now)
	if err != nil {
		return err
	}
	if _, err := s.store.CreateSession(ctx, store.CreateSessionParams{
		UserID:    user.ID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
	}); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		Secure:   strings.HasPrefix(s.cfg.BaseURL, "https://"),
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
	})

	return nil
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		s.logger.Error("write json response", "error", err)
	}
}

func (s *Server) writeError(r *http.Request, w http.ResponseWriter, status int, err error) {
	response := errorResponse{Error: err.Error()}
	if requestID, ok := requestIDFromContext(r.Context()); ok {
		response.RequestID = requestID
	}
	s.writeJSON(w, status, response)
}

func (s *Server) decodeJSON(r *http.Request, value any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func userFromContext(ctx context.Context) (store.User, bool) {
	user, ok := ctx.Value(userContextKey).(store.User)
	return user, ok
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 39 {
		return errors.New("username must be between 3 and 39 characters")
	}
	if !namePattern.MatchString(username) {
		return errors.New("username contains invalid characters")
	}
	return nil
}

func validateRepositoryName(name string) error {
	if len(name) < 1 || len(name) > 100 {
		return errors.New("repository name must be between 1 and 100 characters")
	}
	if !namePattern.MatchString(name) {
		return errors.New("repository name contains invalid characters")
	}
	return nil
}

func normalizeOwnerType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", store.OwnerTypeUser:
		return store.OwnerTypeUser
	case store.OwnerTypeOrganization:
		return store.OwnerTypeOrganization
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func normalizeOrganizationRole(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
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
	switch strings.ToLower(strings.TrimSpace(value)) {
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

func (s *Server) authenticateRequest(r *http.Request) (*store.User, error) {
	if user, err := s.authenticateSession(r); err == nil {
		return user, nil
	}
	if user, err := s.authenticateBasicAuth(r); err == nil {
		return user, nil
	}
	return nil, store.ErrUnauthorized
}

func (s *Server) authenticateSession(r *http.Request) (*store.User, error) {
	cookie, err := r.Cookie(s.cfg.CookieName)
	if err != nil {
		return nil, store.ErrUnauthorized
	}

	claims, err := auth.ParseToken(s.cfg.Secret, cookie.Value)
	if err != nil || claims.ID == "" {
		return nil, store.ErrUnauthorized
	}

	session, err := s.store.GetSessionByTokenID(r.Context(), claims.ID)
	if err != nil {
		return nil, store.ErrUnauthorized
	}
	if session.RevokedAt != nil || session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, store.ErrUnauthorized
	}

	user, err := s.store.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		return nil, store.ErrUnauthorized
	}
	return &user, nil
}

func (s *Server) authenticateBasicAuth(r *http.Request) (*store.User, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, store.ErrUnauthorized
	}
	user, err := s.store.GetUserByUsername(r.Context(), username)
	if err != nil {
		return nil, store.ErrUnauthorized
	}
	if err := auth.CheckPassword(user.PasswordHash, password); err != nil {
		return nil, store.ErrUnauthorized
	}
	return &user, nil
}

func sshFingerprint(publicKey string) (string, error) {
	parsedKey, _, _, _, err := cryptossh.ParseAuthorizedKey([]byte(strings.TrimSpace(publicKey)))
	if err != nil {
		return "", fmt.Errorf("parse public key: %w", err)
	}
	return cryptossh.FingerprintSHA256(parsedKey), nil
}

func newOpaqueID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(raw[:])
}

func normalizeVisibility(visibility string) string {
	if strings.TrimSpace(visibility) == "" {
		return "private"
	}
	return strings.ToLower(strings.TrimSpace(visibility))
}
