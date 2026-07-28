ALTER TABLE exchange_bindings
  DROP CONSTRAINT IF EXISTS exchange_bindings_account_mode_check;

ALTER TABLE exchange_bindings
  ADD CONSTRAINT exchange_bindings_account_mode_check
  CHECK (account_mode IN ('real', 'demo', 'testnet'));
