UPDATE execution_jobs j
SET payload = jsonb_set(
  jsonb_set(j.payload, '{exchange_binding_id}', to_jsonb(j.exchange_binding_id::text), true),
  '{account_mode}',
  to_jsonb(COALESCE(NULLIF(j.payload->>'account_mode', ''), b.account_mode, 'real')),
  true
)
FROM exchange_bindings b
WHERE b.id = j.exchange_binding_id
  AND (
    NOT (j.payload ? 'exchange_binding_id')
    OR j.payload->>'exchange_binding_id' = ''
    OR NOT (j.payload ? 'account_mode')
    OR j.payload->>'account_mode' = ''
  );
