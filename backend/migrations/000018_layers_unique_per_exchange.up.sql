ALTER TABLE layers
  DROP CONSTRAINT IF EXISTS layers_layer_number_unique;

ALTER TABLE layers
  ADD CONSTRAINT layers_layer_number_exchange_unique UNIQUE (user_id, symbol, exchange_binding_id, layer_number);
