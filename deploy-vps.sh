#!/bin/bash
# ============================================
# MAUTRADE — Smart Auto-Deployment Script
# Runs on the VPS after rsync from CI/CD
# Uses docker-compose.prod.yml (Traefik + SSL)
# Usage:
#   CI/CD (automatic): ./deploy-vps.sh
#   Manual:            ./deploy-vps.sh
# ============================================

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_DIR"

COMPOSE_FILE="$PROJECT_DIR/docker-compose.prod.yml"

# ──────────────────────────────────────────────
#  Deploy lock to prevent overlapping deploys
# ──────────────────────────────────────────────
LOCK_DIR="$PROJECT_DIR/.deploy.lock"
if ! mkdir "$LOCK_DIR" 2>/dev/null; then
    echo -e "${RED}Another deployment is already running. Aborting.${NC}"
    exit 1
fi
trap 'rmdir "$LOCK_DIR" 2>/dev/null || true' EXIT

echo -e "${BLUE}==============================================${NC}"
echo -e "${BLUE}  Mautrade Production Deployment${NC}"
echo -e "${BLUE}==============================================${NC}"
echo ""

# ──────────────────────────────────────────────
#  Audit logging
# ──────────────────────────────────────────────
AUDIT_LOG_DIR="$PROJECT_DIR/logs"
AUDIT_LOG_FILE="$AUDIT_LOG_DIR/deploy-audit.log"
mkdir -p "$AUDIT_LOG_DIR"
chmod 700 "$AUDIT_LOG_DIR" 2>/dev/null || true

audit_log() {
    local ACTION="$1"
    local DETAILS="$2"
    local TS
    TS="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
    printf '%s actor=%s source=%s sha=%s action=%s details=%s\n' \
        "$TS" \
        "${GITHUB_ACTOR:-local-admin}" \
        "${GITHUB_WORKFLOW:-manual}" \
        "${GITHUB_SHA:-unknown}" \
        "$ACTION" \
        "$DETAILS" >> "$AUDIT_LOG_FILE"
}

audit_log "deploy_started" "project_dir=$PROJECT_DIR"

# ──────────────────────────────────────────────
#  Helpers
# ──────────────────────────────────────────────
contains_item() {
    local ITEM="$1"
    shift
    for EXISTING in "$@"; do
        [ "$EXISTING" = "$ITEM" ] && return 0
    done
    return 1
}

dc_cmd() {
    if docker compose version &> /dev/null; then
        docker compose -f "$COMPOSE_FILE" "$@"
    elif sudo docker compose version &> /dev/null; then
        sudo docker compose -f "$COMPOSE_FILE" "$@"
    else
        sudo docker-compose -f "$COMPOSE_FILE" "$@"
    fi
}

# ──────────────────────────────────────────────
#  [1/7] Docker check
# ──────────────────────────────────────────────
echo -e "${YELLOW}[1/7] Checking Docker...${NC}"

if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}Docker not found. Installing...${NC}"
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo systemctl enable docker
    sudo systemctl start docker
    sudo usermod -aG docker "$USER" 2>/dev/null || true
    rm get-docker.sh
fi
docker --version
echo -e "${GREEN}  ✓ Docker ready${NC}"

# ──────────────────────────────────────────────
#  [2/7] Detect changed components
# ──────────────────────────────────────────────
echo -e "${YELLOW}[2/7] Detecting changed components...${NC}"

SERVICES_TO_REBUILD=()
REBUILD_ALL=false

CHANGED_FILES_SOURCE=""
if [ -n "${CHANGED_FILES:-}" ]; then
    CHANGED_FILES_SOURCE="env"
elif [ -f ".deploy/changed-files.txt" ]; then
    CHANGED_FILES_SOURCE="file"
else
    REBUILD_ALL=true
    CHANGED_FILES_SOURCE="all"
fi

if [ "$REBUILD_ALL" = true ]; then
    echo -e "${CYAN}  No change list found — rebuilding ALL services${NC}"
    SERVICES_TO_REBUILD=("traefik" "frontend" "api" "postgres" "nats" "keydb" "questdb")
else
    if [ "$CHANGED_FILES_SOURCE" = "env" ]; then
        IFS=' ' read -ra FILES <<< "$CHANGED_FILES"
    else
        mapfile -t FILES < .deploy/changed-files.txt
    fi

    API_CHANGED=false
    FRONTEND_CHANGED=false
    INFRA_CHANGED=false
    RUST_CHANGED=false

    for FILE in "${FILES[@]}"; do
        case "$FILE" in
            backend/cmd/*|backend/internal/*|backend/go.mod|backend/go.sum|backend/Dockerfile|backend/migrations/*)
                API_CHANGED=true
                ;;
            backend/rust-node-execution/*)
                RUST_CHANGED=true
                ;;
            frontend/*)
                FRONTEND_CHANGED=true
                ;;
            docker-compose.prod.yml|deploy-vps.sh|.env.production*)
                INFRA_CHANGED=true
                ;;
        esac
    done

    if [ "$API_CHANGED" = true ]; then
        SERVICES_TO_REBUILD+=("api")
        echo -e "${CYAN}  ✓ Go API changed${NC}"
    fi

    if [ "$RUST_CHANGED" = true ]; then
        echo -e "${CYAN}  ✓ Rust worker changed${NC}"
        if ! contains_item "api" "${SERVICES_TO_REBUILD[@]+"${SERVICES_TO_REBUILD[@]}"}"; then
            SERVICES_TO_REBUILD+=("api")
        fi
    fi

    if [ "$FRONTEND_CHANGED" = true ]; then
        SERVICES_TO_REBUILD+=("frontend")
        echo -e "${CYAN}  ✓ Frontend changed → rebuild Nuxt${NC}"
    fi

    if [ "$INFRA_CHANGED" = true ]; then
        echo -e "${CYAN}  ✓ Infrastructure config changed — rebuilding all${NC}"
        SERVICES_TO_REBUILD=("traefik" "frontend" "api" "postgres" "nats" "keydb" "questdb")
    fi
fi

if [ ${#SERVICES_TO_REBUILD[@]} -eq 0 ]; then
    echo -e "${GREEN}  No services need rebuilding. Skipping Docker operations.${NC}"
    audit_log "deploy_skipped" "reason=no_changes"
fi

# ──────────────────────────────────────────────
#  [3/7] Ensure production .env exists
# ──────────────────────────────────────────────
echo -e "${YELLOW}[3/7] Checking environment config...${NC}"

# Production root .env (used by docker-compose.prod.yml)
if [ ! -f "$PROJECT_DIR/.env" ]; then
    if [ -f "$PROJECT_DIR/.env.production.example" ]; then
        cp "$PROJECT_DIR/.env.production.example" "$PROJECT_DIR/.env"
        echo -e "${YELLOW}  ⚠ Created .env from .env.production.example${NC}"
        echo -e "${YELLOW}    → EDIT .env ON VPS WITH REAL PRODUCTION VALUES!${NC}"
        audit_log "env_created_from_template" "file=.env"
    else
        echo -e "${RED}  ✗ No .env found. Create from .env.production.example${NC}"
    fi
else
    echo -e "${GREEN}  ✓ Root .env exists${NC}"
fi

# Backend .env (used by Go API env_file)
if [ ! -f "$PROJECT_DIR/backend/.env" ]; then
    if [ -f "$PROJECT_DIR/backend/.env.example" ]; then
        cp "$PROJECT_DIR/backend/.env.example" "$PROJECT_DIR/backend/.env"
        sed -i 's/^APP_ENV=.*/APP_ENV=production/' "$PROJECT_DIR/backend/.env"
        # Update CORS for production domain
        DOMAIN=$(grep '^DOMAIN=' "$PROJECT_DIR/.env" 2>/dev/null | cut -d= -f2 || echo "mautrade.com")
        sed -i "s|^ALLOWED_CORS_ORIGIN=.*|ALLOWED_CORS_ORIGIN=https://$DOMAIN|" "$PROJECT_DIR/backend/.env"
        echo -e "${GREEN}  ✓ backend/.env created (production mode)${NC}"
        audit_log "backend_env_created" "mode=production"
    else
        echo -e "${RED}  ✗ No backend/.env.example found${NC}"
    fi
else
    # Safety net: enforce production mode
    if grep -q '^APP_ENV=development' "$PROJECT_DIR/backend/.env" 2>/dev/null; then
        sed -i 's/^APP_ENV=development/APP_ENV=production/' "$PROJECT_DIR/backend/.env"
        echo -e "${YELLOW}  ⚠ Forced APP_ENV=production (was 'development')${NC}"
        audit_log "env_forced_production" "file=backend/.env"
    fi
    # Enforce CORS origin
    DOMAIN=$(grep '^DOMAIN=' "$PROJECT_DIR/.env" 2>/dev/null | cut -d= -f2 || echo "mautrade.com")
    if grep -q 'ALLOWED_CORS_ORIGIN=http://localhost' "$PROJECT_DIR/backend/.env" 2>/dev/null; then
        sed -i "s|^ALLOWED_CORS_ORIGIN=.*|ALLOWED_CORS_ORIGIN=https://$DOMAIN|" "$PROJECT_DIR/backend/.env"
        echo -e "${YELLOW}  ⚠ Forced CORS origin to https://$DOMAIN${NC}"
        audit_log "cors_forced_production" "domain=$DOMAIN"
    fi
    echo -e "${GREEN}  ✓ backend/.env exists (production)${NC}"
fi

# Inject SMTP variables from CI/CD if provided
if [ -n "${SMTP_HOST:-}" ]; then
    for VAR in "SMTP_HOST" "SMTP_PORT" "SMTP_USERNAME" "SMTP_PASSWORD" "SMTP_FROM"; do
        if grep -q "^${VAR}=" "$PROJECT_DIR/backend/.env"; then
            sed -i "s|^${VAR}=.*|${VAR}=${!VAR}|" "$PROJECT_DIR/backend/.env"
        else
            echo "${VAR}=${!VAR}" >> "$PROJECT_DIR/backend/.env"
        fi
    done
    echo -e "${GREEN}  ✓ Injected SMTP configuration into backend/.env${NC}"
    audit_log "smtp_injected" "host=${SMTP_HOST}"
fi

# ──────────────────────────────────────────────
#  [4/7] Build & Deploy Docker services
# ──────────────────────────────────────────────
if [ ${#SERVICES_TO_REBUILD[@]} -gt 0 ]; then
    echo -e "${YELLOW}[4/7] Building and deploying: ${SERVICES_TO_REBUILD[*]}${NC}"

    # Build only services that need building (skip pull-only images)
    BUILDABLE_SERVICES=()
    for SERVICE in "${SERVICES_TO_REBUILD[@]}"; do
        case "$SERVICE" in
            api|frontend)
                BUILDABLE_SERVICES+=("$SERVICE")
                ;;
        esac
    done

    if [ ${#BUILDABLE_SERVICES[@]} -gt 0 ]; then
        echo -e "${BLUE}  Building: ${BUILDABLE_SERVICES[*]}...${NC}"
        dc_cmd build --no-cache "${BUILDABLE_SERVICES[@]}" 2>&1
    fi

    # Pull latest images for infrastructure services
    echo -e "${BLUE}  Pulling latest images...${NC}"
    dc_cmd pull --ignore-buildable 2>/dev/null || true

    # Bring up full stack
    echo -e "${BLUE}  Starting full production stack...${NC}"
    dc_cmd up -d

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}  ✓ Production stack running${NC}"
        audit_log "docker_services_started" "services=${SERVICES_TO_REBUILD[*]}"
    else
        echo -e "${RED}  ✗ Failed to start production stack${NC}"
        audit_log "docker_services_failed" "services=${SERVICES_TO_REBUILD[*]}"
        exit 1
    fi
else
    echo -e "${YELLOW}[4/7] Skipping Docker build (no changes)${NC}"
fi

# ──────────────────────────────────────────────
#  [5/7] Docker cleanup
# ──────────────────────────────────────────────
echo -e "${YELLOW}[5/7] Cleaning up Docker resources...${NC}"

docker image prune -f 2>/dev/null || true
docker builder prune -f --keep-storage=2GB 2>/dev/null || true

DISK_USAGE=$(docker system df --format '{{.TotalCount}} images, {{.Size}}' 2>/dev/null | head -1 || echo "unknown")
echo -e "${GREEN}  ✓ Docker disk usage: $DISK_USAGE${NC}"

audit_log "docker_cleanup_done" "disk=$DISK_USAGE"

# ──────────────────────────────────────────────
#  [6/7] Health checks
# ──────────────────────────────────────────────
echo -e "${YELLOW}[6/7] Running health checks...${NC}"

HEALTH_OK=true
DOMAIN=$(grep '^DOMAIN=' "$PROJECT_DIR/.env" 2>/dev/null | cut -d= -f2 || echo "mautrade.com")

# Check Traefik is up
echo -e "${CYAN}  Checking Traefik (reverse proxy)...${NC}"
for i in $(seq 1 15); do
    if curl -sf -o /dev/null http://localhost:80; then
        echo -e "${GREEN}  ✓ Traefik is responding (${i}s)${NC}"
        break
    fi
    if [ "$i" -eq 15 ]; then
        echo -e "${YELLOW}  ⚠ Traefik not responding on port 80${NC}"
        HEALTH_OK=false
    fi
    sleep 2
done

# Check API via internal port
echo -e "${CYAN}  Checking Go API...${NC}"
for i in $(seq 1 30); do
    # API is internal only, check via docker exec
    if dc_cmd exec -T api wget -q -O /dev/null http://localhost:8080/health 2>/dev/null; then
        echo -e "${GREEN}  ✓ API is healthy (${i}s)${NC}"
        break
    fi
    if [ "$i" -eq 30 ]; then
        echo -e "${YELLOW}  ⚠ API health check inconclusive (may not have /health endpoint yet)${NC}"
    fi
    sleep 1
done

# Check container statuses
echo -e "${CYAN}  Container statuses:${NC}"
dc_cmd ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || dc_cmd ps

# ──────────────────────────────────────────────
#  [7/7] SSL certificate check
# ──────────────────────────────────────────────
echo -e "${YELLOW}[7/7] Checking SSL certificate...${NC}"

# Give Traefik a moment to request the cert
sleep 5

if curl -sf -o /dev/null "https://$DOMAIN" 2>/dev/null; then
    echo -e "${GREEN}  ✓ SSL certificate active for $DOMAIN${NC}"
    audit_log "ssl_check_ok" "domain=$DOMAIN"
else
    echo -e "${YELLOW}  ⚠ SSL not ready yet — Traefik auto-requests from Let's Encrypt${NC}"
    echo -e "${YELLOW}    First-time SSL may take 1-2 minutes. Check: curl -I https://$DOMAIN${NC}"
    audit_log "ssl_check_pending" "domain=$DOMAIN"
fi

# ──────────────────────────────────────────────
#  Done
# ──────────────────────────────────────────────
echo ""
echo -e "${GREEN}==============================================${NC}"
echo -e "${GREEN}  ✓ Mautrade production deployment complete!${NC}"
echo -e "${GREEN}==============================================${NC}"
echo ""
echo -e "  Website:  ${CYAN}https://$DOMAIN${NC}"
echo -e "  API:      ${CYAN}https://$DOMAIN/api/v1${NC}"
echo ""

audit_log "deploy_completed" "health=$HEALTH_OK,domain=$DOMAIN"

# Clean up the CI change list
rm -rf "$PROJECT_DIR/.deploy" 2>/dev/null || true
