UPDATE exchange_bindings
SET status = 'revoked'
WHERE status = 'closing_only';

ALTER TABLE exchange_bindings
  DROP CONSTRAINT IF EXISTS exchange_bindings_status_check;

ALTER TABLE exchange_bindings
  ADD CONSTRAINT exchange_bindings_status_check
  CHECK (status IN ('active', 'invalid', 'revoked'));
