DROP INDEX IF EXISTS idx_gas_fee_deposits_tx_id_status;

ALTER TABLE gas_fee_deposits
  DROP COLUMN IF EXISTS verification_note,
  DROP COLUMN IF EXISTS verified_at,
  DROP COLUMN IF EXISTS actual_amount,
  DROP COLUMN IF EXISTS confirmations,
  DROP COLUMN IF EXISTS block_number,
  DROP COLUMN IF EXISTS sender_address,
  DROP COLUMN IF EXISTS token_contract,
  DROP COLUMN IF EXISTS chain_id,
  DROP COLUMN IF EXISTS network;
