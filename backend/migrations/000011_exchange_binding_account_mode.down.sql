ALTER TABLE exchange_bindings
  DROP CONSTRAINT IF EXISTS exchange_bindings_account_mode_check;

ALTER TABLE exchange_bindings
  DROP COLUMN IF EXISTS account_mode;
