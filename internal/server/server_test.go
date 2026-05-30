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

	"github.com/yashlunawat/forge/internal/config"
	"github.com/yashlunawat/forge/internal/store/memory"
)

func TestAuthAndRepositoryLifecycle(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	app, err := New(
		config.Config{
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
		},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		memory.NewStore(),
	)
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
	app, err := New(
		config.Config{
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
		},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		memory.NewStore(),
	)
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
	app, err := New(
		config.Config{
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
		},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		memory.NewStore(),
	)
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
