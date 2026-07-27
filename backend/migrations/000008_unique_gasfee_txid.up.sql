CREATE UNIQUE INDEX IF NOT EXISTS unique_gas_fee_deposits_tx_id 
ON gas_fee_deposits (tx_id) 
WHERE status IN ('pending', 'confirmed');
