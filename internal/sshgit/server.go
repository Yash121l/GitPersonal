package sshgit

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/yashlunawat/forge/internal/config"
	"github.com/yashlunawat/forge/internal/repository"
	"github.com/yashlunawat/forge/internal/store"
)

type Server struct {
	cfg          config.Config
	logger       *slog.Logger
	store        store.Store
	repositories *repository.Service
	sshConfig    *ssh.ServerConfig
}

func New(cfg config.Config, logger *slog.Logger, st store.Store, repositories *repository.Service) (*Server, error) {
	signer, err := loadOrCreateHostKey(cfg.SSHHostKeyPath)
	if err != nil {
		return nil, err
	}

	server := &Server{
		cfg:          cfg,
		logger:       logger,
		store:        st,
		repositories: repositories,
		sshConfig: &ssh.ServerConfig{
			ServerVersion: "SSH-2.0-Forge",
			PublicKeyCallback: func(metadata ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
				return authenticatePublicKey(st, cfg, metadata, key)
			},
		},
	}
	server.sshConfig.AddHostKey(signer)
	return server, nil
}

func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	var tempDelay sync.Once
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			tempDelay.Do(func() {
				s.logger.Warn("ssh accept failed", "error", err)
			})
			return err
		}

		go s.handleConn(ctx, conn)
	}
}

func (s *Server) handleConn(ctx context.Context, netConn net.Conn) {
	serverConn, channels, requests, err := ssh.NewServerConn(netConn, s.sshConfig)
	if err != nil {
		s.logger.Warn("ssh handshake failed", "error", err)
		return
	}
	defer serverConn.Close()

	go ssh.DiscardRequests(requests)

	for newChannel := range channels {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unsupported channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			s.logger.Warn("accept ssh channel", "error", err)
			continue
		}
		go s.handleSession(ctx, serverConn, channel, requests)
	}
}

func (s *Server) handleSession(ctx context.Context, conn *ssh.ServerConn, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "exec":
			command := parseExecPayload(req.Payload)
			if command == "" {
				_ = req.Reply(false, nil)
				sendExitStatus(channel, 1)
				return
			}

			_ = req.Reply(true, nil)
			if err := s.dispatchExec(ctx, conn, channel, command); err != nil {
				s.logger.Warn("ssh exec failed", "command", command, "error", err)
				sendExitStatus(channel, 1)
				return
			}
			sendExitStatus(channel, 0)
			return
		default:
			_ = req.Reply(false, nil)
		}
	}
}

func (s *Server) dispatchExec(ctx context.Context, conn *ssh.ServerConn, channel ssh.Channel, command string) error {
	gitCommand, owner, repoName, writeAccess, err := parseGitSSHCommand(command)
	if err != nil {
		return err
	}

	user, err := userFromPermissions(s.store, conn.Permissions)
	if err != nil {
		return err
	}

	repositoryMeta, err := s.repositories.GetRepository(ctx, owner, repoName)
	if err != nil {
		return err
	}
	if writeAccess {
		canWrite, err := s.repositories.CanWrite(ctx, &user, repositoryMeta)
		if err != nil {
			return err
		}
		if !canWrite {
			return store.ErrForbidden
		}
	} else {
		canRead, err := s.repositories.CanRead(ctx, &user, repositoryMeta)
		if err != nil {
			return err
		}
		if !canRead {
			return store.ErrForbidden
		}
	}

	cmd := exec.CommandContext(ctx, gitCommand, repositoryMeta.RepoPath)
	cmd.Stdin = channel
	cmd.Stdout = channel
	cmd.Stderr = channel.Stderr()

	if err := cmd.Run(); err != nil {
		return err
	}

	if writeAccess {
		s.repositories.ScheduleMaintenance(repositoryMeta)
	}
	return nil
}

func authenticatePublicKey(st store.Store, cfg config.Config, metadata ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	fingerprint := ssh.FingerprintSHA256(key)
	user, err := st.GetUserBySSHFingerprint(context.Background(), fingerprint)
	if err != nil {
		return nil, err
	}
	if metadata.User() != cfg.SSHUser && !strings.EqualFold(metadata.User(), user.Username) {
		return nil, store.ErrUnauthorized
	}
	_ = st.TouchSSHKeyUsage(context.Background(), fingerprint, time.Now().UTC())
	return &ssh.Permissions{
		Extensions: map[string]string{
			"user_id":     strconv.FormatInt(user.ID, 10),
			"username":    user.Username,
			"user_role":   user.Role,
			"fingerprint": fingerprint,
		},
	}, nil
}

func userFromPermissions(st store.Store, permissions *ssh.Permissions) (store.User, error) {
	if permissions == nil {
		return store.User{}, store.ErrUnauthorized
	}
	rawID := permissions.Extensions["user_id"]
	if rawID == "" {
		return store.User{}, store.ErrUnauthorized
	}
	userID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return store.User{}, err
	}
	return st.GetUserByID(context.Background(), userID)
}

func parseExecPayload(payload []byte) string {
	if len(payload) < 4 {
		return ""
	}
	length := int(uint32(payload[0])<<24 | uint32(payload[1])<<16 | uint32(payload[2])<<8 | uint32(payload[3]))
	if len(payload) < 4+length {
		return ""
	}
	return string(payload[4 : 4+length])
}

func parseGitSSHCommand(command string) (gitCommand, owner, repoName string, writeAccess bool, err error) {
	command = strings.TrimSpace(command)
	fields := strings.Fields(command)
	if len(fields) != 2 {
		return "", "", "", false, errors.New("unsupported ssh command")
	}

	gitCommand = fields[0]
	switch gitCommand {
	case "git-upload-pack":
		writeAccess = false
	case "git-receive-pack":
		writeAccess = true
	default:
		return "", "", "", false, errors.New("unsupported ssh command")
	}

	repoPath := strings.Trim(fields[1], "'\"")
	repoPath = strings.TrimPrefix(repoPath, "/")
	segments := strings.Split(repoPath, "/")
	if len(segments) != 2 || !strings.HasSuffix(segments[1], ".git") {
		return "", "", "", false, errors.New("invalid ssh repository path")
	}
	return gitCommand, segments[0], strings.TrimSuffix(segments[1], ".git"), writeAccess, nil
}

func sendExitStatus(channel ssh.Channel, code uint32) {
	_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{Status: code}))
}

func loadOrCreateHostKey(path string) (ssh.Signer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		_ = publicKey

		privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return nil, err
		}
		block := &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}
		if err := os.WriteFile(path, pem.EncodeToMemory(block), 0o600); err != nil {
			return nil, err
		}
	}

	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(keyBytes)
}
