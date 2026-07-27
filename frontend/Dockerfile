# ── Stage 1: Build ──────────────────────────────
FROM node:22-slim AS build

# Install pnpm directly (avoids corepack download hang in Docker)
RUN npm install -g pnpm@11.13.1

WORKDIR /app

# Install dependencies first (cache layer)
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

# Copy source and build
COPY . .
RUN pnpm build

# ── Stage 2: Runtime ────────────────────────────
FROM node:22-slim
# (Corepack not needed at runtime)

WORKDIR /app

# Copy only the built output
COPY --from=build /app/.output /app/.output

ENV NITRO_PORT=3000
ENV NITRO_HOST=0.0.0.0
ENV NODE_ENV=production

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
