# Mautrade Rust Node Execution

This crate is the Rust worker-side contract for Mautrade execution jobs.

The Go backend publishes idempotent jobs to:

- `execution.buy.request`
- `execution.sell.request`

The Rust worker consumes the `EXECUTION` JetStream stream through a durable pull consumer, routes each job to an exchange adapter, and publishes normalized reports to:

- `execution.result`
- `execution.dlq`

The current adapters are scaffolds only. They intentionally do not include live exchange signing, API-key decryption, exchange filters, or real order placement yet. Those must be implemented against official exchange documentation before production use.
