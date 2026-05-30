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
	"regexp"
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
	Name          string `json:"name"`
	Description   string `json:"description"`
	Visibility    string `json:"visibility"`
	DefaultBranch string `json:"default_branch"`
}

type createSSHKeyRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type currentUserResponse struct {
	User store.User `json:"user"`
}

type repoListResponse struct {
	Repositories []store.Repository `json:"repositories"`
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

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", s.handleRegister)
			r.Post("/login", s.handleLogin)
			r.Post("/logout", s.handleLogout)
		})

		r.With(s.requireAuth).Get("/me", s.handleCurrentUser)
		r.With(s.requireAuth).Post("/keys", s.handleCreateSSHKey)

		r.With(s.requireAuth).Get("/repos", s.handleListRepositories)
		r.With(s.requireAuth).Post("/repos", s.handleCreateRepository)
		r.With(s.requireAuth).Delete("/repos/{owner}/{repo}", s.handleDeleteRepository)
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

	repositories, err := s.store.ListRepositoriesByOwner(r.Context(), user.Username)
	if err != nil {
		s.logger.Error("list repositories", "error", err, "owner", user.Username)
		s.writeError(r, w, http.StatusInternalServerError, errors.New("failed to list repositories"))
		return
	}

	s.writeJSON(w, http.StatusOK, repoListResponse{Repositories: repositories})
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

	repository, err := s.repositories.CreateRepository(r.Context(), store.CreateRepositoryParams{
		Owner:         user.Username,
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
	if !strings.EqualFold(owner, user.Username) && user.Role != "owner" {
		s.writeError(r, w, http.StatusForbidden, errors.New("cannot delete repository owned by another user"))
		return
	}

	if err := s.repositories.DeleteRepository(r.Context(), owner, repo); err != nil {
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
