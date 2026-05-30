CREATE UNIQUE INDEX organizations_slug_lower_idx ON organizations (lower(slug));

CREATE TABLE repo_collaborators (
    repository_id BIGINT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (repository_id, user_id)
);

CREATE INDEX repo_collaborators_user_id_idx ON repo_collaborators (user_id);
