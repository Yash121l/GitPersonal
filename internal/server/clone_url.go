package server

import (
	"net"
	"net/url"
	"strings"
)

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
