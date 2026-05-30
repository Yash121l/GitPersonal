CREATE TABLE repo_webhooks (
    id BIGSERIAL PRIMARY KEY,
    repository_id BIGINT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    secret TEXT NOT NULL DEFAULT '',
    events TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_delivery_at TIMESTAMPTZ,
    last_delivery_status INTEGER,
    last_delivery_error TEXT NOT NULL DEFAULT '',
    success_count BIGINT NOT NULL DEFAULT 0,
    failure_count BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX repo_webhooks_repository_id_idx ON repo_webhooks (repository_id);
