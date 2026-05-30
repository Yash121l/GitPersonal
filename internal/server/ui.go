package server

import (
	"embed"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
)

//go:embed ui/*
var uiAssets embed.FS

var uiStaticFS = mustSubFS(uiAssets, "ui")

var uiPageTemplate = template.Must(template.New("ui-page").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .Title }}</title>
  <link rel="stylesheet" href="/app/assets/app.css">
  <script defer src="/app/assets/app.js"></script>
</head>
<body data-view="{{ .View }}" data-repo-owner="{{ .RepoOwner }}" data-repo-name="{{ .RepoName }}">
  <div class="app-shell">
    <header class="app-topbar">
      <a class="app-brand" href="/app/repos">
        <span class="app-brand-mark">F</span>
        <span class="app-brand-copy">
          <strong>Forge</strong>
          <span>Self-hosted Git</span>
        </span>
      </a>
      <nav class="app-nav">
        <a href="/app/repos">Repositories</a>
        <a href="/app/orgs">Organizations</a>
        <a href="/app/keys">SSH Keys</a>
        <button type="button" class="ghost-button" data-action="logout">Log Out</button>
      </nav>
    </header>
    <main class="app-main">
      <div class="app-messages" data-flash aria-live="polite"></div>
      <div class="app-root" data-app></div>
    </main>
  </div>
</body>
</html>`))

type uiPageData struct {
	Title     string
	View      string
	RepoOwner string
	RepoName  string
}

func (s *Server) handleAppEntry(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticateSession(r); err == nil {
		http.Redirect(w, r, "/app/repos", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/app/login", http.StatusFound)
}

func (s *Server) handleUIPage(title, view string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.renderUIPage(w, uiPageData{Title: title, View: view})
	})
}

func (s *Server) handleUIRepositoryPage(w http.ResponseWriter, r *http.Request) {
	s.renderUIPage(w, uiPageData{
		Title:     "Repository",
		View:      "repo",
		RepoOwner: chi.URLParam(r, "owner"),
		RepoName:  chi.URLParam(r, "repo"),
	})
}

func (s *Server) requireAppSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := s.authenticateSession(r); err != nil {
			http.Redirect(w, r, "/app/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) renderUIPage(w http.ResponseWriter, data uiPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := uiPageTemplate.Execute(w, data); err != nil {
		s.logger.Error("render ui page", "error", err)
	}
}

func (s *Server) cloneURLs(owner, repo string) (string, string) {
	baseURL := strings.TrimSuffix(s.cfg.BaseURL, "/")
	httpCloneURL := baseURL + "/git/" + owner + "/" + repo + ".git"
	if !s.cfg.SSHEnabled {
		return httpCloneURL, ""
	}

	parsedBaseURL, err := url.Parse(s.cfg.BaseURL)
	if err != nil {
		return httpCloneURL, ""
	}

	sshHost := parsedBaseURL.Hostname()
	if sshHost == "" {
		sshHost = "localhost"
	}

	sshPort := ""
	if host, port, err := net.SplitHostPort(s.cfg.SSHAddress); err == nil {
		if host != "" && host != "0.0.0.0" && host != "::" {
			sshHost = host
		}
		sshPort = port
	} else if s.cfg.SSHAddress != "" && !strings.HasPrefix(s.cfg.SSHAddress, ":") {
		sshHost = s.cfg.SSHAddress
	}

	sshCloneURL := "ssh://" + s.cfg.SSHUser + "@" + sshHost
	if sshPort != "" {
		sshCloneURL += ":" + sshPort
	}
	sshCloneURL += "/" + owner + "/" + repo + ".git"

	return httpCloneURL, sshCloneURL
}

func mustSubFS(root fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(root, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
