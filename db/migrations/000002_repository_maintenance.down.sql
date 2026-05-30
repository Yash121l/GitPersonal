ALTER TABLE repositories
    DROP COLUMN IF EXISTS last_maintained_at,
    DROP COLUMN IF EXISTS last_indexed_at;
