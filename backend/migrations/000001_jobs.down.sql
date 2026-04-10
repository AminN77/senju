DROP TRIGGER IF EXISTS trg_jobs_updated_at ON public.jobs;
DROP FUNCTION IF EXISTS public.jobs_set_updated_at();
DROP INDEX IF EXISTS public.idx_jobs_status_active;
DROP INDEX IF EXISTS public.idx_jobs_created_at;
DROP TABLE IF EXISTS public.jobs;
