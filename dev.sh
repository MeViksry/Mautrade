#!/bin/bash
# ============================================
# MAUTRADE — LOCAL DEVELOPMENT SCRIPT
# ============================================
# Run with: chmod +x dev.sh && ./dev.sh
# Commands:
#   ./dev.sh          → Start dev environment
#   ./dev.sh setup    → Full first-time setup
#   ./dev.sh reset    → Wipe everything clean
#   ./dev.sh stop     → Stop all services
#   ./dev.sh logs     → Tail Docker logs
#   ./dev.sh help     → Show help
# ============================================

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$PROJECT_DIR/backend"
FRONTEND_DIR="$PROJECT_DIR/frontend"
RUST_DIR="$BACKEND_DIR/rust-node-execution"

# ============================================
# GENERATE RANDOM SECRET
# ============================================
generate_secret() {
    openssl rand -hex 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 64 | head -n 1
}

# ============================================
# BANNER
# ============================================
show_banner() {
    echo -e "${CYAN}"
    echo "╔═══════════════════════════════════════════╗"
    echo "║         MAUTRADE TRADING PLATFORM         ║"
    echo "║           Development System              ║"
    echo "╚═══════════════════════════════════════════╝"
    echo -e "${NC}"
}

# ============================================
# CHECK IF COMMAND EXISTS
# ============================================
command_exists() {
    command -v "$1" &> /dev/null
}

# Auto-detect docker compose command
dc_cmd() {
    if docker compose version &> /dev/null; then
        docker compose "$@"
    elif sudo docker compose version &> /dev/null; then
        sudo docker compose "$@"
    elif docker-compose --version &> /dev/null; then
        docker-compose "$@"
    else
        sudo docker-compose "$@"
    fi
}

# ============================================
# INSTALL SYSTEM DEPENDENCIES
# ============================================
install_system_deps() {
    echo -e "${BLUE}[1/5] Verifying system dependencies...${NC}"

    # ── Docker ──
    if ! command_exists docker; then
        echo -e "${YELLOW}Docker not found. Installing...${NC}"
        if [ -f /etc/debian_version ] || [ -f /etc/kali_version ]; then
            sudo apt update -y
            sudo apt install -y docker.io docker-compose-plugin
            sudo systemctl enable docker --now 2>/dev/null || sudo service docker start 2>/dev/null
            sudo usermod -aG docker "$USER" 2>/dev/null || true
        else
            curl -fsSL https://get.docker.com -o get-docker.sh
            sudo sh get-docker.sh
            sudo systemctl enable docker --now
            sudo usermod -aG docker "$USER" 2>/dev/null || true
            rm get-docker.sh
        fi
    fi
    echo -e "${GREEN}  ✓ Docker $(docker --version | grep -oP '\d+\.\d+\.\d+')${NC}"

    # ── Node.js ──
    if ! command_exists node; then
        echo -e "${YELLOW}Node.js not found. Installing via nvm...${NC}"
        export NVM_DIR="$HOME/.nvm"
        if [ ! -s "$NVM_DIR/nvm.sh" ]; then
            curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
        fi
        \. "$NVM_DIR/nvm.sh"
        nvm install 22
        nvm use 22
        nvm alias default 22
    else
        # Source nvm if it exists (for consistent node version)
        export NVM_DIR="$HOME/.nvm"
        [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
    fi
    echo -e "${GREEN}  ✓ Node.js $(node -v)${NC}"

    # ── pnpm ──
    if ! command_exists pnpm; then
        echo -e "${YELLOW}pnpm not found. Installing...${NC}"
        npm install -g pnpm
    fi
    echo -e "${GREEN}  ✓ pnpm $(pnpm -v)${NC}"

    # ── Go ──
    if command_exists go; then
        echo -e "${GREEN}  ✓ Go $(go version | grep -oP 'go\d+\.\d+\.\d+')${NC}"
    else
        echo -e "${YELLOW}  ⚠ Go not installed (needed only if running API outside Docker)${NC}"
    fi

    # ── Rust ──
    if command_exists cargo; then
        echo -e "${GREEN}  ✓ Rust $(rustc --version | grep -oP '\d+\.\d+\.\d+')${NC}"
    else
        echo -e "${YELLOW}  ⚠ Rust not installed (needed only if building rust-node-execution locally)${NC}"
    fi

    # ── Essential tools ──
    if [ -f /etc/debian_version ] || [ -f /etc/kali_version ]; then
        sudo apt install -y curl openssl git 2>/dev/null || true
    fi

    echo -e "${GREEN}  ✓ System dependencies verified${NC}"
}

# ============================================
# SETUP ENVIRONMENT VARIABLES
# ============================================
setup_environments() {
    echo -e "${BLUE}[2/5] Configuring environment variables...${NC}"

    # ── Backend .env ──
    if [ ! -f "$BACKEND_DIR/.env" ]; then
        EXCHANGE_KEY=$(generate_secret)

        cat > "$BACKEND_DIR/.env" <<EOF
APP_ENV=development
HTTP_ADDR=:8080
DATABASE_URL=postgres://mautrade:mautrade@localhost:5432/mautrade?sslmode=disable
NATS_URL=nats://localhost:4222
GAS_FEE_SHARE_RATE=0.5
DEFAULT_CURRENCY=USDT
GAS_FEE_DEPOSIT_ADDRESS=MAUTRADE-USDT-DEPOSIT-PENDING
EXCHANGE_CREDENTIAL_KEY=$EXCHANGE_KEY
ALLOWED_CORS_ORIGIN=http://localhost:3000
SHUTDOWN_TIMEOUT_SECONDS=15
AUTH_SESSION_TTL_HOURS=720
EMAIL_OTP_TTL_MINUTES=10
ADMIN_BOOTSTRAP_EMAIL=
ADMIN_BOOTSTRAP_PASSWORD=
ADMIN_BOOTSTRAP_NAME=
ADMIN_BOOTSTRAP_ROLE=

# SMTP Configuration for Email Verification
SMTP_HOST=
SMTP_PORT=
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=
EOF
        echo -e "${GREEN}  ✓ backend/.env created${NC}"
    else
        echo -e "${CYAN}  ● backend/.env already exists (skipped)${NC}"
    fi

    # ── Rust worker .env ──
    if [ ! -f "$RUST_DIR/.env" ]; then
        cat > "$RUST_DIR/.env" <<EOF
APP_ENV=development
NATS_URL=nats://localhost:4222
DATABASE_URL=postgres://mautrade:mautrade@localhost:5432/mautrade?sslmode=disable
KEYDB_ADDR=localhost:6379
QUESTDB_ADDR=localhost:8812
EOF
        echo -e "${GREEN}  ✓ rust-node-execution/.env created${NC}"
    else
        echo -e "${CYAN}  ● rust-node-execution/.env already exists (skipped)${NC}"
    fi

    # ── Frontend .env ──
    if [ ! -f "$FRONTEND_DIR/.env" ]; then
        cat > "$FRONTEND_DIR/.env" <<EOF
NUXT_PUBLIC_API_BASE=http://localhost:8080/api/v1
EOF
        echo -e "${GREEN}  ✓ frontend/.env created${NC}"
    else
        echo -e "${CYAN}  ● frontend/.env already exists (skipped)${NC}"
    fi
}

# ============================================
# INSTALL FRONTEND DEPENDENCIES
# ============================================
install_frontend_deps() {
    echo -e "${BLUE}[3/5] Installing frontend dependencies...${NC}"
    if [ -d "$FRONTEND_DIR" ]; then
        cd "$FRONTEND_DIR"
        pnpm install
        cd "$PROJECT_DIR"
        echo -e "${GREEN}  ✓ Frontend dependencies installed${NC}"
    else
        echo -e "${YELLOW}  ⚠ Frontend directory not found${NC}"
    fi
}

# ============================================
# START INFRASTRUCTURE (Docker Compose)
# ============================================
start_infrastructure() {
    echo -e "${BLUE}[4/5] Starting infrastructure (Postgres, NATS, KeyDB, QuestDB)...${NC}"

    if ! docker info &> /dev/null; then
        echo -e "${RED}Docker daemon is not running!${NC}"
        echo -e "${YELLOW}Try: sudo systemctl start docker${NC}"
        exit 1
    fi

    cd "$BACKEND_DIR"
    dc_cmd up --build -d

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}  ✓ Docker services are running${NC}"
    else
        echo -e "${RED}  ✗ Failed to start Docker services${NC}"
        exit 1
    fi

    cd "$PROJECT_DIR"

    # Wait for Postgres health check
    echo -e "${CYAN}  Waiting for Postgres to be ready...${NC}"
    for i in $(seq 1 30); do
        if cd "$BACKEND_DIR" && dc_cmd exec -T postgres pg_isready -U mautrade -d mautrade > /dev/null 2>&1; then
            echo -e "${GREEN}  ✓ Postgres is ready (${i}s)${NC}"
            cd "$PROJECT_DIR"
            break
        fi
        cd "$PROJECT_DIR"
        if [ "$i" -eq 30 ]; then
            echo -e "${YELLOW}  ⚠ Postgres may still be initializing...${NC}"
        fi
        sleep 1
    done
}

# ============================================
# START DEV ENVIRONMENT
# ============================================
start_dev() {
    show_banner
    echo -e "${GREEN}Starting Mautrade Development Environment${NC}"
    echo "=================================================="
    echo ""

    # Start infrastructure
    start_infrastructure

    # Install frontend deps if missing
    if [ -d "$FRONTEND_DIR" ] && [ ! -d "$FRONTEND_DIR/node_modules" ]; then
        echo -e "${YELLOW}Frontend modules missing. Installing...${NC}"
        cd "$FRONTEND_DIR" && pnpm install && cd "$PROJECT_DIR"
    fi

    # Start Nuxt frontend in dev mode
    FRONTEND_PID=""
    if [ -d "$FRONTEND_DIR" ] && [ -f "$FRONTEND_DIR/package.json" ]; then
        echo -e "${BLUE}Starting Nuxt frontend (dev mode)...${NC}"
        cd "$FRONTEND_DIR"
        pnpm dev &
        FRONTEND_PID=$!
        cd "$PROJECT_DIR"
    fi

    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║         MAUTRADE IS RUNNING! 🚀              ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  Frontend (Nuxt):  ${CYAN}http://localhost:3000${NC}"
    echo -e "  Backend API:      ${CYAN}http://localhost:8080${NC}"
    echo -e "  NATS Monitor:     ${CYAN}http://localhost:8222${NC}"
    echo -e "  QuestDB Console:  ${CYAN}http://localhost:9000${NC}"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop frontend. Docker services keep running.${NC}"
    echo -e "${YELLOW}Use './dev.sh stop' to stop everything.${NC}"
    echo ""

    # Graceful shutdown on Ctrl+C
    cleanup() {
        echo -e "\n${YELLOW}Stopping frontend...${NC}"
        [ -n "$FRONTEND_PID" ] && kill "$FRONTEND_PID" 2>/dev/null
        echo -e "${GREEN}Frontend stopped. Docker services still running.${NC}"
        echo -e "${CYAN}Run './dev.sh stop' to shut down Docker too.${NC}"
        exit 0
    }
    trap cleanup SIGINT SIGTERM

    wait "$FRONTEND_PID" 2>/dev/null
}

# ============================================
# FULL SETUP (FIRST TIME)
# ============================================
full_setup() {
    show_banner
    echo -e "${GREEN}Running Full First-Time Setup...${NC}"
    echo "=================================================="
    echo ""

    install_system_deps
    setup_environments
    install_frontend_deps
    start_infrastructure

    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║             SETUP COMPLETE! ✅               ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "Run ${YELLOW}./dev.sh${NC} to spin up the dev environment."
    echo ""
    echo -e "${CYAN}Docker Services:${NC}"
    echo -e "  Postgres:  ${CYAN}localhost:5432${NC}  (user: mautrade / pass: mautrade)"
    echo -e "  NATS:      ${CYAN}localhost:4222${NC}  (JetStream enabled)"
    echo -e "  KeyDB:     ${CYAN}localhost:6379${NC}"
    echo -e "  QuestDB:   ${CYAN}localhost:9000${NC}  (Web UI) / ${CYAN}:8812${NC} (PG wire)"
    echo -e "  Go API:    ${CYAN}localhost:8080${NC}  (via Docker)"
    echo ""
}

# ============================================
# STOP ALL SERVICES
# ============================================
stop_all() {
    echo -e "${YELLOW}Stopping all Docker services...${NC}"
    cd "$BACKEND_DIR"
    dc_cmd down
    cd "$PROJECT_DIR"
    echo -e "${GREEN}All services stopped.${NC}"
}

# ============================================
# SHOW LOGS
# ============================================
show_logs() {
    cd "$BACKEND_DIR"
    dc_cmd logs -f --tail=100 "$@"
    cd "$PROJECT_DIR"
}

# ============================================
# RESET (DELETE ALL AND RESTART)
# ============================================
reset_all() {
    echo -e "${RED}╔══════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║  WARNING: This will DESTROY everything!      ║${NC}"
    echo -e "${RED}║  - All Docker containers & volumes            ║${NC}"
    echo -e "${RED}║  - All .env files                             ║${NC}"
    echo -e "${RED}║  - Frontend node_modules                      ║${NC}"
    echo -e "${RED}╚══════════════════════════════════════════════╝${NC}"
    echo ""
    read -p "Are you sure? (y/N): " CONFIRM

    if [[ "$CONFIRM" =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Stopping and removing Docker containers + volumes...${NC}"
        cd "$BACKEND_DIR"
        dc_cmd down -v 2>/dev/null || true
        cd "$PROJECT_DIR"

        echo -e "${YELLOW}Removing frontend build artifacts...${NC}"
        rm -rf "$FRONTEND_DIR/node_modules"
        rm -rf "$FRONTEND_DIR/.nuxt"
        rm -rf "$FRONTEND_DIR/.output"
        rm -rf "$FRONTEND_DIR/.nitro"

        echo -e "${YELLOW}Removing Rust build artifacts...${NC}"
        rm -rf "$RUST_DIR/target"

        echo -e "${YELLOW}Removing .env files...${NC}"
        rm -f "$BACKEND_DIR/.env"
        rm -f "$RUST_DIR/.env"
        rm -f "$FRONTEND_DIR/.env"

        echo -e "${YELLOW}Removing deploy artifacts...${NC}"
        rm -rf "$PROJECT_DIR/.deploy"
        rm -rf "$PROJECT_DIR/.deploy.lock"
        rm -rf "$PROJECT_DIR/logs"

        echo ""
        echo -e "${GREEN}Hard reset complete. Run './dev.sh setup' to reinstall.${NC}"
    else
        echo -e "${CYAN}Reset cancelled.${NC}"
    fi
}

# ============================================
# STATUS CHECK
# ============================================
show_status() {
    show_banner
    echo -e "${BLUE}Service Status:${NC}"
    echo ""

    cd "$BACKEND_DIR"
    dc_cmd ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || dc_cmd ps
    cd "$PROJECT_DIR"

    echo ""

    # Check .env files
    echo -e "${BLUE}Environment Files:${NC}"
    [ -f "$BACKEND_DIR/.env" ] && echo -e "  ${GREEN}✓${NC} backend/.env" || echo -e "  ${RED}✗${NC} backend/.env (missing)"
    [ -f "$RUST_DIR/.env" ] && echo -e "  ${GREEN}✓${NC} rust-node-execution/.env" || echo -e "  ${RED}✗${NC} rust-node-execution/.env (missing)"
    [ -f "$FRONTEND_DIR/.env" ] && echo -e "  ${GREEN}✓${NC} frontend/.env" || echo -e "  ${RED}✗${NC} frontend/.env (missing)"

    echo ""

    # Check frontend
    echo -e "${BLUE}Frontend:${NC}"
    [ -d "$FRONTEND_DIR/node_modules" ] && echo -e "  ${GREEN}✓${NC} node_modules installed" || echo -e "  ${RED}✗${NC} node_modules missing (run ./dev.sh setup)"

    echo ""
}

# ============================================
# SHOW HELP
# ============================================
show_help() {
    show_banner
    echo "Usage: ./dev.sh [command]"
    echo ""
    echo "Commands:"
    echo -e "  ${GREEN}(none)${NC}     Start dev environment (frontend + Docker stack)"
    echo -e "  ${GREEN}setup${NC}      Full first-time installation & configuration"
    echo -e "  ${GREEN}stop${NC}       Stop all Docker services"
    echo -e "  ${GREEN}status${NC}     Show service statuses"
    echo -e "  ${GREEN}logs${NC}       Tail Docker logs (optionally: ./dev.sh logs api)"
    echo -e "  ${GREEN}reset${NC}      Wipe everything clean (Docker volumes, .envs, deps)"
    echo -e "  ${GREEN}help${NC}       Show this menu"
    echo ""
    echo "Examples:"
    echo "  ./dev.sh setup          # First-time setup"
    echo "  ./dev.sh                # Start developing"
    echo "  ./dev.sh logs api       # Follow Go API logs"
    echo "  ./dev.sh logs postgres  # Follow Postgres logs"
    echo "  ./dev.sh stop           # Stop everything"
    echo ""
}

# ============================================
# MAIN
# ============================================
case "${1:-}" in
    setup)
        full_setup
        ;;
    stop)
        stop_all
        ;;
    status)
        show_status
        ;;
    logs)
        shift
        show_logs "$@"
        ;;
    reset)
        reset_all
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        if [ ! -f "$BACKEND_DIR/.env" ]; then
            echo -e "${YELLOW}First time? Running full setup...${NC}"
            full_setup
            echo ""
            echo -e "${GREEN}Setup done! Starting dev environment...${NC}"
            echo ""
        fi
        start_dev
        ;;
esac
