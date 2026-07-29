ALTER TABLE layers
  DROP CONSTRAINT IF EXISTS layers_layer_number_exchange_unique;

ALTER TABLE layers
  ADD CONSTRAINT layers_layer_number_unique UNIQUE (user_id, symbol, layer_number);
