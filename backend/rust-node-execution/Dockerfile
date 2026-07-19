# ── Stage 1: Build ──────────────────────────────
FROM rust:1.87-bookworm AS build

WORKDIR /src
COPY Cargo.toml Cargo.lock ./

# Pre-build dependencies (cache layer)
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release && rm -rf src

# Copy real source and build
COPY src/ src/
RUN touch src/main.rs && cargo build --release --bin mautrade-rust-node-execution

# ── Stage 2: Runtime ────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r mautrade && useradd -r -g mautrade mautrade

COPY --from=build /src/target/release/mautrade-rust-node-execution /usr/local/bin/mautrade-rust-node-execution

USER mautrade

ENTRYPOINT ["mautrade-rust-node-execution"]
