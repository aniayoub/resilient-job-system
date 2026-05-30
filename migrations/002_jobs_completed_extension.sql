TRANSACTION START;
ALTER TABLE jobs
ADD COLUMN completed_at TIMESTAMPTZ;

UPDATE jobs SET completed_at = updated_at
WHERE status = 'completed';

COMMIT;