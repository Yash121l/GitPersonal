package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yashlunawat/forge/internal/config"
	"github.com/yashlunawat/forge/internal/database"
	"github.com/yashlunawat/forge/internal/store"
)

func TestStoreRoundTrip(t *testing.T) {
	t.Parallel()

	testDB := newTestDatabase(t)
	st := newTestStore(t, testDB)

	if err := st.Check(context.Background()); err != nil {
		t.Fatalf("check store: %v", err)
	}

	user, err := st.CreateUser(context.Background(), "yash", "hash", "member")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := st.CreateUser(context.Background(), "YASH", "hash", "member"); !errors.Is(err, store.ErrAlreadyExists) {
		t.Fatalf("expected duplicate user create to fail with already exists, got %v", err)
	}

	gotUser, err := st.GetUserByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	if gotUser.Username != "yash" {
		t.Fatalf("unexpected username: %s", gotUser.Username)
	}

	gotByUsername, err := st.GetUserByUsername(context.Background(), "YASH")
	if err != nil {
		t.Fatalf("get user by username: %v", err)
	}
	if gotByUsername.ID != user.ID {
		t.Fatalf("unexpected user id from username lookup: got %d want %d", gotByUsername.ID, user.ID)
	}

	repository, err := st.CreateRepository(context.Background(), store.CreateRepositoryParams{
		Owner:         "yash",
		Name:          "forge",
		Description:   "Self-hosted git platform",
		Visibility:    "private",
		DefaultBranch: "main",
		RepoPath:      "/data/repos/forge.git",
	})
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}
	if _, err := st.CreateRepository(context.Background(), store.CreateRepositoryParams{
		Owner:         "yash",
		Name:          "FORGE",
		Description:   "Duplicate",
		Visibility:    "private",
		DefaultBranch: "main",
		RepoPath:      "/data/repos/forge-duplicate.git",
	}); !errors.Is(err, store.ErrAlreadyExists) {
		t.Fatalf("expected duplicate repository create to fail with already exists, got %v", err)
	}

	gotRepository, err := st.GetRepositoryByOwnerAndName(context.Background(), "YASH", "FORGE")
	if err != nil {
		t.Fatalf("get repository: %v", err)
	}
	if gotRepository.ID != repository.ID {
		t.Fatalf("unexpected repository id: got %d want %d", gotRepository.ID, repository.ID)
	}

	repositories, err := st.ListRepositories(context.Background())
	if err != nil {
		t.Fatalf("list repositories: %v", err)
	}
	if len(repositories) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(repositories))
	}

	ownedRepositories, err := st.ListRepositoriesByOwner(context.Background(), "yash")
	if err != nil {
		t.Fatalf("list repositories by owner: %v", err)
	}
	if len(ownedRepositories) != 1 {
		t.Fatalf("expected 1 owned repository, got %d", len(ownedRepositories))
	}

	indexedAt := time.Now().UTC().Round(time.Microsecond)
	maintainedAt := indexedAt.Add(time.Minute)
	if err := st.UpdateRepositoryStats(context.Background(), "yash", "forge", 4096, &indexedAt, &maintainedAt); err != nil {
		t.Fatalf("update repository stats: %v", err)
	}

	updatedRepository, err := st.GetRepositoryByOwnerAndName(context.Background(), "yash", "forge")
	if err != nil {
		t.Fatalf("get updated repository: %v", err)
	}
	if updatedRepository.SizeBytes != 4096 {
		t.Fatalf("unexpected repository size: got %d", updatedRepository.SizeBytes)
	}
	if updatedRepository.LastIndexedAt == nil || updatedRepository.LastMaintainedAt == nil {
		t.Fatalf("expected maintenance timestamps to be set: %+v", updatedRepository)
	}

	session, err := st.CreateSession(context.Background(), store.CreateSessionParams{
		UserID:    user.ID,
		TokenID:   "11111111-1111-1111-1111-111111111111",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := st.CreateSession(context.Background(), store.CreateSessionParams{
		UserID:    user.ID,
		TokenID:   session.TokenID,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}); !errors.Is(err, store.ErrAlreadyExists) {
		t.Fatalf("expected duplicate session create to fail with already exists, got %v", err)
	}

	gotSession, err := st.GetSessionByTokenID(context.Background(), session.TokenID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if gotSession.ID != session.ID {
		t.Fatalf("unexpected session id: got %d want %d", gotSession.ID, session.ID)
	}

	revokedAt := time.Now().UTC().Round(time.Microsecond)
	if err := st.RevokeSession(context.Background(), session.TokenID, revokedAt); err != nil {
		t.Fatalf("revoke session: %v", err)
	}

	revokedSession, err := st.GetSessionByTokenID(context.Background(), session.TokenID)
	if err != nil {
		t.Fatalf("get revoked session: %v", err)
	}
	if revokedSession.RevokedAt == nil {
		t.Fatal("expected revoked session to record revoked_at")
	}

	key, err := st.CreateSSHKey(context.Background(), store.CreateSSHKeyParams{
		UserID:            user.ID,
		Name:              "laptop",
		PublicKey:         "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBtestkey",
		FingerprintSHA256: "SHA256:test",
	})
	if err != nil {
		t.Fatalf("create ssh key: %v", err)
	}
	if _, err := st.CreateSSHKey(context.Background(), store.CreateSSHKeyParams{
		UserID:            user.ID,
		Name:              "duplicate",
		PublicKey:         key.PublicKey,
		FingerprintSHA256: key.FingerprintSHA256,
	}); !errors.Is(err, store.ErrAlreadyExists) {
		t.Fatalf("expected duplicate ssh key create to fail with already exists, got %v", err)
	}

	if key.UserID != user.ID {
		t.Fatalf("unexpected ssh key owner: got %d want %d", key.UserID, user.ID)
	}

	userByFingerprint, err := st.GetUserBySSHFingerprint(context.Background(), key.FingerprintSHA256)
	if err != nil {
		t.Fatalf("get user by ssh fingerprint: %v", err)
	}
	if userByFingerprint.ID != user.ID {
		t.Fatalf("unexpected ssh fingerprint lookup result: got %d want %d", userByFingerprint.ID, user.ID)
	}

	if err := st.TouchSSHKeyUsage(context.Background(), key.FingerprintSHA256, time.Now().UTC()); err != nil {
		t.Fatalf("touch ssh key usage: %v", err)
	}

	if err := st.DeleteRepository(context.Background(), "yash", "forge"); err != nil {
		t.Fatalf("delete repository: %v", err)
	}

	if _, err := st.GetRepositoryByOwnerAndName(context.Background(), "yash", "forge"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected deleted repository lookup to fail with not found, got %v", err)
	}
}

func TestWithRepositoryLeaseSerializesSameRepository(t *testing.T) {
	t.Parallel()

	testDB := newTestDatabase(t)
	primary := newTestStore(t, testDB)
	secondary := newTestStore(t, testDB)

	acquired := make(chan struct{})
	release := make(chan struct{})
	secondEntered := make(chan struct{})
	errs := make(chan error, 2)

	go func() {
		errs <- primary.WithRepositoryLease(context.Background(), "yash", "forge", func(context.Context) error {
			close(acquired)
			<-release
			return nil
		})
	}()

	waitForSignal(t, acquired, time.Second, "first lease to be acquired")

	go func() {
		errs <- secondary.WithRepositoryLease(context.Background(), "yash", "forge", func(context.Context) error {
			close(secondEntered)
			return nil
		})
	}()

	assertNoSignal(t, secondEntered, 250*time.Millisecond, "second lease to stay blocked")
	close(release)
	waitForSignal(t, secondEntered, time.Second, "second lease to acquire after release")
	assertNoError(t, <-errs)
	assertNoError(t, <-errs)
}

func TestRepositoryWebhookRoundTrip(t *testing.T) {
	t.Parallel()

	testDB := newTestDatabase(t)
	st := newTestStore(t, testDB)

	if _, err := st.CreateUser(context.Background(), "yash", "hash", "member"); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := st.CreateRepository(context.Background(), store.CreateRepositoryParams{
		Owner:         "yash",
		Name:          "forge",
		Description:   "Self-hosted git platform",
		Visibility:    "private",
		DefaultBranch: "main",
		RepoPath:      "/data/repos/forge.git",
	}); err != nil {
		t.Fatalf("create repository: %v", err)
	}

	webhook, err := st.CreateRepositoryWebhook(context.Background(), store.CreateRepositoryWebhookParams{
		Owner:    "yash",
		RepoName: "forge",
		URL:      "https://hooks.example.test/forge",
		Secret:   "top-secret",
		Events:   []string{store.RepositoryWebhookEventPush, store.RepositoryWebhookEventDeleted},
	})
	if err != nil {
		t.Fatalf("create repository webhook: %v", err)
	}

	webhooks, err := st.ListRepositoryWebhooks(context.Background(), "yash", "forge")
	if err != nil {
		t.Fatalf("list repository webhooks: %v", err)
	}
	if len(webhooks) != 1 {
		t.Fatalf("expected 1 repository webhook, got %+v", webhooks)
	}
	if webhooks[0].ID != webhook.ID || webhooks[0].URL != webhook.URL {
		t.Fatalf("unexpected repository webhook: %+v", webhooks[0])
	}

	deliveredAt := time.Now().UTC().Round(time.Microsecond)
	if err := st.RecordRepositoryWebhookDelivery(context.Background(), store.RecordRepositoryWebhookDeliveryParams{
		WebhookID:   webhook.ID,
		DeliveredAt: deliveredAt,
		StatusCode:  204,
	}); err != nil {
		t.Fatalf("record repository webhook delivery: %v", err)
	}

	webhooks, err = st.ListRepositoryWebhooks(context.Background(), "yash", "forge")
	if err != nil {
		t.Fatalf("list repository webhooks after delivery: %v", err)
	}
	if webhooks[0].LastDeliveryAt == nil || webhooks[0].LastDeliveryStatus != 204 {
		t.Fatalf("expected delivery status to be recorded, got %+v", webhooks[0])
	}
	if webhooks[0].SuccessCount != 1 || webhooks[0].FailureCount != 0 {
		t.Fatalf("unexpected delivery counters: %+v", webhooks[0])
	}

	if err := st.DeleteRepositoryWebhook(context.Background(), "yash", "forge", webhook.ID); err != nil {
		t.Fatalf("delete repository webhook: %v", err)
	}

	webhooks, err = st.ListRepositoryWebhooks(context.Background(), "yash", "forge")
	if err != nil {
		t.Fatalf("list repository webhooks after delete: %v", err)
	}
	if len(webhooks) != 0 {
		t.Fatalf("expected 0 repository webhooks after delete, got %+v", webhooks)
	}
}

func TestOrganizationAndCollaboratorRoundTrip(t *testing.T) {
	t.Parallel()

	testDB := newTestDatabase(t)
	st := newTestStore(t, testDB)

	alice, err := st.CreateUser(context.Background(), "alice", "hash", "member")
	if err != nil {
		t.Fatalf("create alice: %v", err)
	}
	bob, err := st.CreateUser(context.Background(), "bob", "hash", "member")
	if err != nil {
		t.Fatalf("create bob: %v", err)
	}
	carol, err := st.CreateUser(context.Background(), "carol", "hash", "member")
	if err != nil {
		t.Fatalf("create carol: %v", err)
	}

	organization, err := st.CreateOrganization(context.Background(), store.CreateOrganizationParams{
		Slug:        "team",
		DisplayName: "Team",
		Description: "shared ownership",
		CreatedBy:   alice.ID,
	})
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}
	if _, err := st.CreateOrganization(context.Background(), store.CreateOrganizationParams{
		Slug:        "alice",
		DisplayName: "Collision",
		Description: "invalid",
		CreatedBy:   alice.ID,
	}); !errors.Is(err, store.ErrAlreadyExists) {
		t.Fatalf("expected org slug collision with username to fail, got %v", err)
	}
	if _, err := st.CreateUser(context.Background(), "TEAM", "hash", "member"); !errors.Is(err, store.ErrAlreadyExists) {
		t.Fatalf("expected username collision with org slug to fail, got %v", err)
	}

	gotOrganization, err := st.GetOrganizationBySlug(context.Background(), "TEAM")
	if err != nil {
		t.Fatalf("get organization: %v", err)
	}
	if gotOrganization.ID != organization.ID {
		t.Fatalf("unexpected organization id: got %d want %d", gotOrganization.ID, organization.ID)
	}

	ownerMembership, err := st.GetOrganizationMembership(context.Background(), "team", alice.ID)
	if err != nil {
		t.Fatalf("get owner membership: %v", err)
	}
	if ownerMembership.Role != store.OrganizationRoleOwner {
		t.Fatalf("expected org creator to be owner, got %s", ownerMembership.Role)
	}

	addedMembership, err := st.AddOrganizationMember(context.Background(), store.AddOrganizationMemberParams{
		OrganizationSlug: "team",
		Username:         "bob",
		Role:             store.OrganizationRoleMaintainer,
	})
	if err != nil {
		t.Fatalf("add bob membership: %v", err)
	}
	if addedMembership.Role != store.OrganizationRoleMaintainer {
		t.Fatalf("unexpected membership role: %s", addedMembership.Role)
	}

	memberships, err := st.ListOrganizationsByMember(context.Background(), bob.ID)
	if err != nil {
		t.Fatalf("list bob memberships: %v", err)
	}
	if len(memberships) != 1 || memberships[0].OrganizationSlug != "team" {
		t.Fatalf("unexpected org memberships: %+v", memberships)
	}

	repository, err := st.CreateRepository(context.Background(), store.CreateRepositoryParams{
		Owner:         "team",
		OwnerType:     store.OwnerTypeOrganization,
		Name:          "infra",
		Description:   "org repo",
		Visibility:    "private",
		DefaultBranch: "main",
		RepoPath:      "/data/repos/team/infra.git",
	})
	if err != nil {
		t.Fatalf("create org repository: %v", err)
	}
	if repository.OwnerType != store.OwnerTypeOrganization {
		t.Fatalf("expected org repo owner type, got %s", repository.OwnerType)
	}

	collaborator, err := st.AddRepositoryCollaborator(context.Background(), store.AddRepositoryCollaboratorParams{
		Owner:    "team",
		RepoName: "infra",
		Username: "carol",
		Role:     store.RepositoryRoleWrite,
	})
	if err != nil {
		t.Fatalf("add collaborator: %v", err)
	}
	if collaborator.Role != store.RepositoryRoleWrite {
		t.Fatalf("unexpected collaborator role: %s", collaborator.Role)
	}

	gotCollaborator, err := st.GetRepositoryCollaborator(context.Background(), "team", "infra", carol.ID)
	if err != nil {
		t.Fatalf("get collaborator: %v", err)
	}
	if gotCollaborator.Username != "carol" {
		t.Fatalf("unexpected collaborator username: %s", gotCollaborator.Username)
	}

	bobRepos, err := st.ListRepositoriesForUser(context.Background(), bob.ID)
	if err != nil {
		t.Fatalf("list bob repos: %v", err)
	}
	if len(bobRepos) != 1 || bobRepos[0].Owner != "team" || bobRepos[0].OwnerType != store.OwnerTypeOrganization {
		t.Fatalf("unexpected bob accessible repos: %+v", bobRepos)
	}

	carolRepos, err := st.ListRepositoriesForUser(context.Background(), carol.ID)
	if err != nil {
		t.Fatalf("list carol repos: %v", err)
	}
	if len(carolRepos) != 1 || carolRepos[0].Name != "infra" {
		t.Fatalf("unexpected carol accessible repos: %+v", carolRepos)
	}
}

func TestWithRepositoryLeaseDoesNotBlockDifferentRepositories(t *testing.T) {
	t.Parallel()

	testDB := newTestDatabase(t)
	primary := newTestStore(t, testDB)
	secondary := newTestStore(t, testDB)

	acquired := make(chan struct{})
	release := make(chan struct{})
	secondEntered := make(chan struct{})
	errs := make(chan error, 2)

	go func() {
		errs <- primary.WithRepositoryLease(context.Background(), "yash", "forge", func(context.Context) error {
			close(acquired)
			<-release
			return nil
		})
	}()

	waitForSignal(t, acquired, time.Second, "first lease to be acquired")

	go func() {
		errs <- secondary.WithRepositoryLease(context.Background(), "yash", "other", func(context.Context) error {
			close(secondEntered)
			return nil
		})
	}()

	waitForSignal(t, secondEntered, time.Second, "different repository lease to proceed")
	close(release)
	assertNoError(t, <-errs)
	assertNoError(t, <-errs)
}

func TestWithRepositoryLeaseReleasesOnCallbackError(t *testing.T) {
	t.Parallel()

	testDB := newTestDatabase(t)
	primary := newTestStore(t, testDB)
	secondary := newTestStore(t, testDB)
	expectedErr := errors.New("boom")

	err := primary.WithRepositoryLease(context.Background(), "yash", "forge", func(context.Context) error {
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected callback error %v, got %v", expectedErr, err)
	}

	acquired := make(chan struct{})
	errs := make(chan error, 1)
	go func() {
		errs <- secondary.WithRepositoryLease(context.Background(), "yash", "forge", func(context.Context) error {
			close(acquired)
			return nil
		})
	}()

	waitForSignal(t, acquired, time.Second, "lease to be reacquired after callback error")
	assertNoError(t, <-errs)
}

type testDatabase struct {
	databaseURL string
}

func newTestDatabase(t *testing.T) testDatabase {
	t.Helper()

	baseURL := os.Getenv("FORGE_TEST_DATABASE_URL")
	if baseURL == "" {
		t.Skip("FORGE_TEST_DATABASE_URL is not set")
	}

	adminDB := openPostgres(t, baseURL)
	dbName := fmt.Sprintf("forge_test_%d", time.Now().UnixNano())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", quoteIdentifier(dbName))); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create test database: %v", err)
	}

	databaseURL := replaceDatabaseName(t, baseURL, dbName)
	testDB := openPostgres(t, databaseURL)
	if err := database.Migrate(context.Background(), testDB); err != nil {
		_ = testDB.Close()
		_ = adminDB.Close()
		t.Fatalf("migrate test database: %v", err)
	}
	if err := testDB.Close(); err != nil {
		_ = adminDB.Close()
		t.Fatalf("close migrated test database: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()

		if _, err := adminDB.ExecContext(cleanupCtx, `
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = $1
  AND pid <> pg_backend_pid()`, dbName); err != nil {
			t.Errorf("terminate test database backends: %v", err)
		}

		if _, err := adminDB.ExecContext(cleanupCtx, fmt.Sprintf("DROP DATABASE %s", quoteIdentifier(dbName))); err != nil {
			t.Errorf("drop test database: %v", err)
		}

		if err := adminDB.Close(); err != nil {
			t.Errorf("close admin database: %v", err)
		}
	})

	return testDatabase{databaseURL: databaseURL}
}

func newTestStore(t *testing.T, testDB testDatabase) *Store {
	t.Helper()

	db := openPostgres(t, testDB.databaseURL)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	return NewStore(db)
}

func openPostgres(t *testing.T, databaseURL string) *sql.DB {
	t.Helper()

	cfg := config.Config{
		DatabaseURL:       databaseURL,
		DBMaxOpenConns:    10,
		DBMaxIdleConns:    10,
		DBConnMaxLifetime: time.Hour,
		DBConnMaxIdleTime: time.Minute,
	}

	db, err := database.OpenPostgres(context.Background(), cfg)
	if err != nil {
		t.Fatalf("open postgres database: %v", err)
	}

	return db
}

func replaceDatabaseName(t *testing.T, rawURL, dbName string) string {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse database url: %v", err)
	}

	parsed.Path = "/" + dbName
	return parsed.String()
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func waitForSignal(t *testing.T, ch <-chan struct{}, timeout time.Duration, description string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func assertNoSignal(t *testing.T, ch <-chan struct{}, timeout time.Duration, description string) {
	t.Helper()

	select {
	case <-ch:
		t.Fatalf("did not expect %s", description)
	case <-time.After(timeout):
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
