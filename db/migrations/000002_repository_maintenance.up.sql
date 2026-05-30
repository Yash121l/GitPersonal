ALTER TABLE repositories
    ADD COLUMN last_indexed_at TIMESTAMPTZ,
    ADD COLUMN last_maintained_at TIMESTAMPTZ;

