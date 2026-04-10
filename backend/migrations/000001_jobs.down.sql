DROP TRIGGER IF EXISTS trg_jobs_updated_at ON jobs;
DROP FUNCTION IF EXISTS jobs_set_updated_at();
DROP INDEX IF EXISTS idx_jobs_status_active;
DROP INDEX IF EXISTS idx_jobs_created_at;
DROP TABLE IF EXISTS jobs;
DROP EXTENSION IF EXISTS pgcrypto;
