package repository

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/yashlunawat/forge/internal/store"
)

type WebhookDispatcher struct {
	logger *slog.Logger
	store  store.Store
	client *http.Client
	queue  chan webhookDeliveryJob
}

type webhookDeliveryJob struct {
	webhook store.RepositoryWebhook
	payload repositoryWebhookPayload
}

type repositoryWebhookPayload struct {
	Event      string                  `json:"event"`
	DeliveryID string                  `json:"delivery_id"`
	OccurredAt time.Time               `json:"occurred_at"`
	Repository store.Repository        `json:"repository"`
	Actor      *repositoryWebhookActor `json:"actor,omitempty"`
}

type repositoryWebhookActor struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func NewWebhookDispatcher(logger *slog.Logger, st store.Store) *WebhookDispatcher {
	return &WebhookDispatcher{
		logger: logger,
		store:  st,
		client: &http.Client{Timeout: 5 * time.Second},
		queue:  make(chan webhookDeliveryJob, 256),
	}
}

func (d *WebhookDispatcher) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-d.queue:
				d.deliver(ctx, job)
			}
		}
	}()
}

func (d *WebhookDispatcher) EnqueueRepositoryEvent(ctx context.Context, repository store.Repository, event string, actor *store.User) {
	webhooks, err := d.store.ListRepositoryWebhooks(ctx, repository.Owner, repository.Name)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			d.logger.Warn("list repository webhooks", "owner", repository.Owner, "repo", repository.Name, "event", event, "error", err)
		}
		return
	}

	d.EnqueueRepositoryEventWithHooks(repository, event, actor, webhooks)
}

func (d *WebhookDispatcher) EnqueueRepositoryEventWithHooks(repository store.Repository, event string, actor *store.User, webhooks []store.RepositoryWebhook) {
	payload := repositoryWebhookPayload{
		Event:      event,
		OccurredAt: time.Now().UTC(),
		Repository: repository,
	}
	if actor != nil {
		payload.Actor = &repositoryWebhookActor{
			ID:       actor.ID,
			Username: actor.Username,
			Role:     actor.Role,
		}
	}

	for _, webhook := range webhooks {
		if !webhookMatchesEvent(webhook, event) {
			continue
		}
		job := webhookDeliveryJob{
			webhook: webhook,
			payload: payload,
		}
		job.payload.DeliveryID = newDeliveryID()
		select {
		case d.queue <- job:
		default:
			d.logger.Warn("repository webhook queue full", "webhook_id", webhook.ID, "owner", repository.Owner, "repo", repository.Name, "event", event)
		}
	}
}

func (d *WebhookDispatcher) deliver(ctx context.Context, job webhookDeliveryJob) {
	body, err := json.Marshal(job.payload)
	if err != nil {
		d.logger.Error("marshal webhook payload", "webhook_id", job.webhook.ID, "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, job.webhook.URL, bytes.NewReader(body))
	if err != nil {
		d.logger.Warn("build webhook request", "webhook_id", job.webhook.ID, "error", err)
		d.recordDelivery(job.webhook.ID, 0, err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Forge-Webhooks/1.0")
	req.Header.Set("X-Forge-Event", job.payload.Event)
	req.Header.Set("X-Forge-Delivery", job.payload.DeliveryID)
	if job.webhook.Secret != "" {
		req.Header.Set("X-Forge-Signature-256", "sha256="+signWebhookPayload(job.webhook.Secret, body))
	}

	response, err := d.client.Do(req)
	if err != nil {
		d.logger.Warn("deliver webhook", "webhook_id", job.webhook.ID, "url", job.webhook.URL, "error", err)
		d.recordDelivery(job.webhook.ID, 0, err.Error())
		return
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)

	errorMessage := ""
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		errorMessage = response.Status
		d.logger.Warn("webhook returned non-success status", "webhook_id", job.webhook.ID, "status", response.StatusCode)
	}
	d.recordDelivery(job.webhook.ID, response.StatusCode, errorMessage)
}

func (d *WebhookDispatcher) recordDelivery(webhookID int64, statusCode int, errorMessage string) {
	err := d.store.RecordRepositoryWebhookDelivery(context.Background(), store.RecordRepositoryWebhookDeliveryParams{
		WebhookID:   webhookID,
		DeliveredAt: time.Now().UTC(),
		StatusCode:  statusCode,
		Error:       errorMessage,
	})
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		d.logger.Warn("record webhook delivery", "webhook_id", webhookID, "error", err)
	}
}

func webhookMatchesEvent(webhook store.RepositoryWebhook, event string) bool {
	for _, candidate := range webhook.Events {
		if candidate == event {
			return true
		}
	}
	return false
}

func signWebhookPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func newDeliveryID() string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(raw[:])
}
