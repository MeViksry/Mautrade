ALTER TABLE exchange_bindings
  ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ;

UPDATE exchange_bindings
SET status = CASE
  WHEN lower(status) IN ('active', 'connected', 'connect') THEN 'active'
  WHEN lower(status) IN ('revoked', 'disconnected', 'disconnect', 'deleted', 'delete') THEN 'revoked'
  WHEN lower(status) = 'invalid' THEN 'invalid'
  ELSE 'invalid'
END;

ALTER TABLE exchange_bindings
  DROP CONSTRAINT IF EXISTS exchange_bindings_status_check;

ALTER TABLE exchange_bindings
  ADD CONSTRAINT exchange_bindings_status_check
  CHECK (status IN ('active', 'invalid', 'revoked'));
