UPDATE layers
SET 
  remaining_quantity = remaining_quantity - sub.total_fee,
  entry_quantity = entry_quantity - sub.total_fee
FROM (
  SELECT layer_id, COALESCE(SUM(exchange_fee), 0) as total_fee
  FROM layer_executions
  WHERE action = 'buy' AND status IN ('success', 'partial')
  GROUP BY layer_id
) sub
WHERE layers.id = sub.layer_id 
  AND layers.status IN ('open', 'partial')
  AND layers.entry_quantity > sub.total_fee;
