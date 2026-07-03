DROP INDEX IF EXISTS u_idx_urls_original_url;
CREATE UNIQUE INDEX IF NOT EXISTS u_idx_urls_original_url ON urls (original_url, deleted_at) NULLS NOT DISTINCT;
