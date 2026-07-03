UPDATE urls SET deleted_at = NOW() WHERE is_deleted = TRUE AND deleted_at IS NULL;
