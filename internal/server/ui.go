package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"slices"
)

//go:embed all:ui/dist
var uiAssets embed.FS

var uiStaticFS = mustSubFS(uiAssets, "ui/dist")
var uiBundle = mustLoadUIBundle(uiStaticFS)

var uiBootstrap = mustMarshalUIBootstrap(map[string]any{
	"basePath":    "/app/",
	"productName": "Forge",
	"features": map[string]bool{
		"workspaceOverview":    true,
		"repositories":         true,
		"organizations":        true,
		"sshKeys":              true,
		"repositoryCode":       true,
		"repositoryAccess":     true,
		"repositoryAutomation": true,
		"repositoryActivity":   true,
		"repositorySettings":   true,
	},
})

var uiPageTemplate = template.Must(template.New("ui-page").Parse(`<!DOCTYPE html>
<html lang="en" class="dark">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .Title }}</title>
  <meta name="description" content="Forge is a self-hosted Git platform for trusted teams.">
  <meta name="theme-color" content="#09090b">
  {{- range .Stylesheets }}
  <link rel="stylesheet" href="/app/assets/{{ . }}">
  {{- end }}
</head>
<body>
  <div id="app"></div>
  <script>
    window.__FORGE_BOOTSTRAP__ = {{ .Bootstrap }};
  </script>
  <script type="module" src="/app/assets/{{ .EntryScript }}"></script>
</body>
</html>`))

type uiPageData struct {
	Title       string
	Bootstrap   template.JS
	EntryScript string
	Stylesheets []string
}

type uiManifestEntry struct {
	File    string   `json:"file"`
	IsEntry bool     `json:"isEntry"`
	CSS     []string `json:"css"`
}

type uiBundleAssets struct {
	EntryScript string
	Stylesheets []string
}

func (s *Server) handleAppEntry(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticateSession(r); err == nil {
		http.Redirect(w, r, "/app/overview", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/app/login", http.StatusFound)
}

func (s *Server) handleUIPage(title string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.renderUIPage(w, uiPageData{Title: title})
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
	if data.Bootstrap == "" {
		data.Bootstrap = uiBootstrap
	}
	if data.EntryScript == "" {
		data.EntryScript = uiBundle.EntryScript
	}
	if len(data.Stylesheets) == 0 {
		data.Stylesheets = uiBundle.Stylesheets
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := uiPageTemplate.Execute(w, data); err != nil {
		s.logger.Error("render ui page", "error", err)
	}
}

func mustSubFS(root fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(root, dir)
	if err != nil {
		panic(err)
	}
	return sub
}

func mustMarshalUIBootstrap(payload any) template.JS {
	encoded, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return template.JS(encoded)
}

func mustLoadUIBundle(root fs.FS) uiBundleAssets {
	manifestBytes, err := fs.ReadFile(root, "manifest.json")
	if err != nil {
		panic(fmt.Sprintf("read vite manifest: %v", err))
	}

	var manifest map[string]uiManifestEntry
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		panic(fmt.Sprintf("decode vite manifest: %v", err))
	}

	for _, key := range []string{"index.html", "src/main.ts"} {
		entry, ok := manifest[key]
		if !ok || !entry.IsEntry || entry.File == "" {
			continue
		}

		return uiBundleAssets{
			EntryScript: entry.File,
			Stylesheets: resolveManifestStylesheets(manifest, entry),
		}
	}

	for _, entry := range manifest {
		if !entry.IsEntry || entry.File == "" {
			continue
		}
		return uiBundleAssets{
			EntryScript: entry.File,
			Stylesheets: resolveManifestStylesheets(manifest, entry),
		}
	}

	panic("vite manifest did not contain an entry asset")
}

func resolveManifestStylesheets(manifest map[string]uiManifestEntry, entry uiManifestEntry) []string {
	if len(entry.CSS) > 0 {
		stylesheets := slices.Clone(entry.CSS)
		slices.Sort(stylesheets)
		return stylesheets
	}

	var stylesheets []string
	for _, candidate := range manifest {
		if candidate.File == "" || path.Ext(candidate.File) != ".css" {
			continue
		}
		stylesheets = append(stylesheets, candidate.File)
	}
	slices.Sort(stylesheets)
	return stylesheets
}

var hashedAssetPattern = regexp.MustCompile(`-[A-Za-z0-9_-]{6,}\.`)

func uiAssetCacheControl(name string) string {
	base := path.Base(name)
	if hashedAssetPattern.MatchString(base) {
		return "public, max-age=31536000, immutable"
	}
	return "no-cache"
}

func (s *Server) handleUIAssets() http.Handler {
	fileServer := http.StripPrefix("/app/assets/", http.FileServer(http.FS(uiStaticFS)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", uiAssetCacheControl(path.Base(r.URL.Path)))
		fileServer.ServeHTTP(w, r)
	})
}
