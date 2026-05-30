package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cryptossh "golang.org/x/crypto/ssh"

	"github.com/yashlunawat/forge/internal/config"
	"github.com/yashlunawat/forge/internal/repository"
	"github.com/yashlunawat/forge/internal/store/memory"
)

func TestAuthAndRepositoryLifecycle(t *testing.T) {
	t.Parallel()

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

	registerBody := map[string]string{
		"username": "yash",
		"password": "correct horse battery staple",
	}
	registerRecorder := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/auth/register", registerBody, nil)
	if registerRecorder.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", registerRecorder.Code, registerRecorder.Body.String())
	}

	cookie := firstCookie(t, registerRecorder.Result().Cookies(), "forge_session")

	createRepoBody := map[string]string{
		"name":           "forge",
		"description":    "Self-hosted git platform",
		"visibility":     "private",
		"default_branch": "main",
	}
	createRepoRecorder := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/repos", createRepoBody, cookie)
	if createRepoRecorder.Code != http.StatusCreated {
		t.Fatalf("create repo status = %d, body = %s", createRepoRecorder.Code, createRepoRecorder.Body.String())
	}

	listRecorder := performJSONRequest(t, app.Router(), http.MethodGet, "/api/v1/repos", nil, cookie)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list repos status = %d, body = %s", listRecorder.Code, listRecorder.Body.String())
	}

	var listBody struct {
		Repositories []struct {
			Name  string `json:"name"`
			Owner string `json:"owner"`
		} `json:"repositories"`
	}
	if err := json.Unmarshal(listRecorder.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listBody.Repositories) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(listBody.Repositories))
	}
	if listBody.Repositories[0].Name != "forge" || listBody.Repositories[0].Owner != "yash" {
		t.Fatalf("unexpected repository payload: %+v", listBody.Repositories[0])
	}

	deleteRecorder := performJSONRequest(t, app.Router(), http.MethodDelete, "/api/v1/repos/yash/forge", nil, cookie)
	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("delete repo status = %d, body = %s", deleteRecorder.Code, deleteRecorder.Body.String())
	}
}

func TestUnauthorizedRepositoryAccess(t *testing.T) {
	t.Parallel()

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

	recorder := performJSONRequest(t, app.Router(), http.MethodGet, "/api/v1/repos", nil, nil)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestReadyzAndSecurityHeaders(t *testing.T) {
	t.Parallel()

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

	recorder := performJSONRequest(t, app.Router(), http.MethodGet, "/readyz", nil, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("missing security header")
	}
	if recorder.Header().Get("X-Request-Id") == "" {
		t.Fatalf("missing request id header")
	}
}

func TestOrganizationAndCollaboratorLifecycle(t *testing.T) {
	t.Parallel()

	app, _ := newTestServer(t)

	aliceRegister := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "alice",
		"password": "correct horse battery staple",
	}, nil)
	if aliceRegister.Code != http.StatusCreated {
		t.Fatalf("alice register status = %d, body = %s", aliceRegister.Code, aliceRegister.Body.String())
	}
	aliceCookie := firstCookie(t, aliceRegister.Result().Cookies(), "forge_session")

	bobRegister := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "bob",
		"password": "correct horse battery staple",
	}, nil)
	if bobRegister.Code != http.StatusCreated {
		t.Fatalf("bob register status = %d, body = %s", bobRegister.Code, bobRegister.Body.String())
	}
	bobCookie := firstCookie(t, bobRegister.Result().Cookies(), "forge_session")

	createOrg := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/orgs", map[string]string{
		"slug":         "team",
		"display_name": "Team",
		"description":  "shared ownership",
	}, aliceCookie)
	if createOrg.Code != http.StatusCreated {
		t.Fatalf("create org status = %d, body = %s", createOrg.Code, createOrg.Body.String())
	}

	addMember := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/orgs/team/members", map[string]string{
		"username": "bob",
		"role":     "maintainer",
	}, aliceCookie)
	if addMember.Code != http.StatusCreated {
		t.Fatalf("add member status = %d, body = %s", addMember.Code, addMember.Body.String())
	}

	createOrgRepo := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/repos", map[string]string{
		"owner":          "team",
		"owner_type":     "organization",
		"name":           "infra",
		"description":    "team repo",
		"visibility":     "private",
		"default_branch": "main",
	}, bobCookie)
	if createOrgRepo.Code != http.StatusCreated {
		t.Fatalf("create org repo status = %d, body = %s", createOrgRepo.Code, createOrgRepo.Body.String())
	}

	createPersonalRepo := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/repos", map[string]string{
		"name":           "personal",
		"description":    "alice repo",
		"visibility":     "private",
		"default_branch": "main",
	}, aliceCookie)
	if createPersonalRepo.Code != http.StatusCreated {
		t.Fatalf("create personal repo status = %d, body = %s", createPersonalRepo.Code, createPersonalRepo.Body.String())
	}

	addCollaborator := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/repos/alice/personal/collaborators", map[string]string{
		"username": "bob",
		"role":     "write",
	}, aliceCookie)
	if addCollaborator.Code != http.StatusCreated {
		t.Fatalf("add collaborator status = %d, body = %s", addCollaborator.Code, addCollaborator.Body.String())
	}

	listRepos := performJSONRequest(t, app.Router(), http.MethodGet, "/api/v1/repos", nil, bobCookie)
	if listRepos.Code != http.StatusOK {
		t.Fatalf("list repos status = %d, body = %s", listRepos.Code, listRepos.Body.String())
	}

	var listBody struct {
		Repositories []struct {
			Owner     string `json:"owner"`
			OwnerType string `json:"owner_type"`
			Name      string `json:"name"`
		} `json:"repositories"`
	}
	if err := json.Unmarshal(listRepos.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode repo list: %v", err)
	}
	if len(listBody.Repositories) != 2 {
		t.Fatalf("expected 2 accessible repositories for bob, got %+v", listBody.Repositories)
	}

	seen := map[string]bool{}
	for _, repository := range listBody.Repositories {
		seen[repository.OwnerType+":"+repository.Owner+"/"+repository.Name] = true
	}
	if !seen["organization:team/infra"] {
		t.Fatalf("expected bob to see org repository, got %+v", listBody.Repositories)
	}
	if !seen["user:alice/personal"] {
		t.Fatalf("expected bob to see collaborator repository, got %+v", listBody.Repositories)
	}

	listOrgs := performJSONRequest(t, app.Router(), http.MethodGet, "/api/v1/orgs", nil, bobCookie)
	if listOrgs.Code != http.StatusOK {
		t.Fatalf("list orgs status = %d, body = %s", listOrgs.Code, listOrgs.Body.String())
	}

	deletePersonalAsBob := performJSONRequest(t, app.Router(), http.MethodDelete, "/api/v1/repos/alice/personal", nil, bobCookie)
	if deletePersonalAsBob.Code != http.StatusForbidden {
		t.Fatalf("expected collaborator write delete to be forbidden, got %d with body %s", deletePersonalAsBob.Code, deletePersonalAsBob.Body.String())
	}

	deleteOrgAsBob := performJSONRequest(t, app.Router(), http.MethodDelete, "/api/v1/repos/team/infra", nil, bobCookie)
	if deleteOrgAsBob.Code != http.StatusForbidden {
		t.Fatalf("expected org maintainer delete to be forbidden, got %d with body %s", deleteOrgAsBob.Code, deleteOrgAsBob.Body.String())
	}
}

func TestRepositoryDetailsSSHKeysAndUIRoutes(t *testing.T) {
	t.Parallel()

	app, _ := newTestServer(t)

	register := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "yash",
		"password": "correct horse battery staple",
	}, nil)
	if register.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", register.Code, register.Body.String())
	}
	cookie := firstCookie(t, register.Result().Cookies(), "forge_session")

	createRepo := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/repos", map[string]string{
		"name":           "forge",
		"description":    "Self-hosted git platform",
		"visibility":     "private",
		"default_branch": "main",
	}, cookie)
	if createRepo.Code != http.StatusCreated {
		t.Fatalf("create repo status = %d, body = %s", createRepo.Code, createRepo.Body.String())
	}

	sshPublicKey, _ := generateRSAKey(t)
	addKey := performJSONRequest(t, app.Router(), http.MethodPost, "/api/v1/keys", map[string]string{
		"name":       "laptop",
		"public_key": string(cryptossh.MarshalAuthorizedKey(sshPublicKey)),
	}, cookie)
	if addKey.Code != http.StatusCreated {
		t.Fatalf("add key status = %d, body = %s", addKey.Code, addKey.Body.String())
	}

	listKeys := performJSONRequest(t, app.Router(), http.MethodGet, "/api/v1/keys", nil, cookie)
	if listKeys.Code != http.StatusOK {
		t.Fatalf("list keys status = %d, body = %s", listKeys.Code, listKeys.Body.String())
	}

	var keysBody struct {
		Keys []struct {
			Name string `json:"name"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(listKeys.Body.Bytes(), &keysBody); err != nil {
		t.Fatalf("decode keys response: %v", err)
	}
	if len(keysBody.Keys) != 1 || keysBody.Keys[0].Name != "laptop" {
		t.Fatalf("unexpected keys payload: %+v", keysBody.Keys)
	}

	repoDetail := performJSONRequest(t, app.Router(), http.MethodGet, "/api/v1/repos/yash/forge", nil, cookie)
	if repoDetail.Code != http.StatusOK {
		t.Fatalf("repo detail status = %d, body = %s", repoDetail.Code, repoDetail.Body.String())
	}

	var detailBody struct {
		HTTPCloneURL string `json:"http_clone_url"`
		SSHCloneURL  string `json:"ssh_clone_url"`
	}
	if err := json.Unmarshal(repoDetail.Body.Bytes(), &detailBody); err != nil {
		t.Fatalf("decode repo detail response: %v", err)
	}
	if detailBody.HTTPCloneURL == "" {
		t.Fatal("expected http clone url to be present")
	}
	if detailBody.SSHCloneURL == "" {
		t.Fatal("expected ssh clone url to be present")
	}

	loginPage := performJSONRequest(t, app.Router(), http.MethodGet, "/app/login", nil, nil)
	if loginPage.Code != http.StatusOK || !bytes.Contains(loginPage.Body.Bytes(), []byte(`data-view="login"`)) {
		t.Fatalf("unexpected login page response: %d %s", loginPage.Code, loginPage.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/app/repos", nil)
	redirectRecorder := httptest.NewRecorder()
	app.Router().ServeHTTP(redirectRecorder, req)
	if redirectRecorder.Code != http.StatusFound {
		t.Fatalf("expected unauthenticated app repos request to redirect, got %d", redirectRecorder.Code)
	}
	if location := redirectRecorder.Header().Get("Location"); location != "/app/login" {
		t.Fatalf("unexpected redirect location: %s", location)
	}

	reposPageRequest := httptest.NewRequest(http.MethodGet, "/app/repos", nil)
	reposPageRequest.AddCookie(cookie)
	reposPage := httptest.NewRecorder()
	app.Router().ServeHTTP(reposPage, reposPageRequest)
	if reposPage.Code != http.StatusOK || !bytes.Contains(reposPage.Body.Bytes(), []byte(`data-view="repos"`)) {
		t.Fatalf("unexpected repos page response: %d %s", reposPage.Code, reposPage.Body.String())
	}

	repoPageRequest := httptest.NewRequest(http.MethodGet, "/app/repos/yash/forge", nil)
	repoPageRequest.AddCookie(cookie)
	repoPage := httptest.NewRecorder()
	app.Router().ServeHTTP(repoPage, repoPageRequest)
	if repoPage.Code != http.StatusOK || !bytes.Contains(repoPage.Body.Bytes(), []byte(`data-view="repo"`)) {
		t.Fatalf("unexpected repo page response: %d %s", repoPage.Code, repoPage.Body.String())
	}
	if !bytes.Contains(repoPage.Body.Bytes(), []byte(`data-repo-owner="yash"`)) {
		t.Fatalf("expected repo page owner metadata, got %s", repoPage.Body.String())
	}

	assetRequest := httptest.NewRequest(http.MethodGet, "/app/assets/app.css", nil)
	assetResponse := httptest.NewRecorder()
	app.Router().ServeHTTP(assetResponse, assetRequest)
	if assetResponse.Code != http.StatusOK {
		t.Fatalf("expected app css asset to be served, got %d", assetResponse.Code)
	}
	if contentType := assetResponse.Header().Get("Content-Type"); contentType == "" || !bytes.Contains([]byte(contentType), []byte("text/css")) {
		t.Fatalf("unexpected app css content type: %s", contentType)
	}
	if !bytes.Contains(assetResponse.Body.Bytes(), []byte("--bg:")) {
		t.Fatalf("expected app css payload, got %s", assetResponse.Body.String())
	}
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func firstCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}

	t.Fatalf("missing cookie %q", name)
	return nil
}
