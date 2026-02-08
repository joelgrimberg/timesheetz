#!/bin/bash
#
# Timesheetz PostgreSQL Setup Script
#
# This script creates the necessary files to run PostgreSQL in Docker
# for use with Timesheetz sync functionality.
#
# Usage:
#   ./setup-postgres.sh [OPTIONS]
#
# Options:
#   -d, --dir DIR       Directory to create setup files (default: ./timesheetz-db)
#   -h, --host HOST     Host/IP where PostgreSQL will be accessible (default: localhost)
#   -p, --port PORT     Port for PostgreSQL (default: 5432)
#   --password PASS     Use specific password (default: auto-generated)
#   --no-start          Don't start the container, just create files
#   --help              Show this help message

set -e

# Default values
INSTALL_DIR="./timesheetz-db"
PG_HOST="localhost"
PG_PORT="5432"
PG_PASSWORD=""
NO_START=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[OK]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

show_help() {
    sed -n '3,18p' "$0" | sed 's/^# //' | sed 's/^#//'
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -h|--host)
            PG_HOST="$2"
            shift 2
            ;;
        -p|--port)
            PG_PORT="$2"
            shift 2
            ;;
        --password)
            PG_PASSWORD="$2"
            shift 2
            ;;
        --no-start)
            NO_START=true
            shift
            ;;
        --help)
            show_help
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Generate password if not provided
if [ -z "$PG_PASSWORD" ]; then
    if command -v openssl &> /dev/null; then
        PG_PASSWORD=$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)
    else
        PG_PASSWORD=$(head -c 32 /dev/urandom | base64 | tr -d '/+=' | head -c 32)
    fi
fi

print_info "Timesheetz PostgreSQL Setup"
echo ""

# Check for Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    echo "Please install Docker first: https://docs.docker.com/get-docker/"
    exit 1
fi
print_success "Docker found"

# Check for docker-compose or docker compose
COMPOSE_CMD=""
if command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
else
    print_error "Docker Compose is not installed"
    echo "Please install Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi
print_success "Docker Compose found ($COMPOSE_CMD)"

# Create directory
print_info "Creating directory: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

# Create docker-compose.yml
cat > "$INSTALL_DIR/docker-compose.yml" << EOF
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: timesheetz-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: timesheetz
      POSTGRES_USER: timesheetz
      POSTGRES_PASSWORD: ${PG_PASSWORD}
    volumes:
      - ./data:/var/lib/postgresql/data
    ports:
      - "${PG_PORT}:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U timesheetz -d timesheetz"]
      interval: 10s
      timeout: 5s
      retries: 5
EOF
print_success "Created docker-compose.yml"

# Create .env file for reference
cat > "$INSTALL_DIR/.env" << EOF
# Timesheetz PostgreSQL Configuration
# Generated: $(date)

POSTGRES_HOST=${PG_HOST}
POSTGRES_PORT=${PG_PORT}
POSTGRES_DB=timesheetz
POSTGRES_USER=timesheetz
POSTGRES_PASSWORD=${PG_PASSWORD}

# Connection URL for Timesheetz
TIMESHEETZ_POSTGRES_URL=postgres://timesheetz:${PG_PASSWORD}@${PG_HOST}:${PG_PORT}/timesheetz?sslmode=disable
EOF
print_success "Created .env file with credentials"

# Create README
cat > "$INSTALL_DIR/README.md" << 'EOF'
# Timesheetz PostgreSQL Database

## Quick Start

```bash
# Start the database
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f

# Stop the database
docker-compose down
```

## Configuration

Add the following to your Timesheetz config (`~/.config/timesheetz/config.yaml`):

```yaml
postgresURL: "<see .env file for URL>"
```

Or set the environment variable:

```bash
export TIMESHEETZ_POSTGRES_URL="<see .env file for URL>"
```

## Backup

```bash
docker exec timesheetz-postgres pg_dump -U timesheetz timesheetz > backup.sql
```

## Restore

```bash
cat backup.sql | docker exec -i timesheetz-postgres psql -U timesheetz timesheetz
```
EOF
print_success "Created README.md"

# Build the connection URL
CONNECTION_URL="postgres://timesheetz:${PG_PASSWORD}@${PG_HOST}:${PG_PORT}/timesheetz?sslmode=disable"

echo ""
print_info "Setup files created in: $INSTALL_DIR"
echo ""

# Start container if requested
if [ "$NO_START" = false ]; then
    print_info "Starting PostgreSQL container..."
    cd "$INSTALL_DIR"
    $COMPOSE_CMD up -d
    
    # Wait for PostgreSQL to be ready
    print_info "Waiting for PostgreSQL to be ready..."
    for i in {1..30}; do
        if docker exec timesheetz-postgres pg_isready -U timesheetz -d timesheetz &> /dev/null; then
            print_success "PostgreSQL is ready!"
            break
        fi
        sleep 1
        if [ $i -eq 30 ]; then
            print_warning "PostgreSQL may still be starting. Check with: docker-compose logs"
        fi
    done
    cd - > /dev/null
fi

echo ""
echo "=============================================="
echo -e "${GREEN}Setup Complete!${NC}"
echo "=============================================="
echo ""
echo "Connection URL:"
echo -e "  ${YELLOW}${CONNECTION_URL}${NC}"
echo ""
echo "Add to ~/.config/timesheetz/config.yaml:"
echo -e "  ${BLUE}postgresURL: \"${CONNECTION_URL}\"${NC}"
echo ""
echo "Or set environment variable:"
echo -e "  ${BLUE}export TIMESHEETZ_POSTGRES_URL=\"${CONNECTION_URL}\"${NC}"
echo ""
echo "Files created:"
echo "  $INSTALL_DIR/docker-compose.yml"
echo "  $INSTALL_DIR/.env (contains credentials)"
echo "  $INSTALL_DIR/README.md"
echo ""

if [ "$NO_START" = true ]; then
    echo "To start PostgreSQL:"
    echo "  cd $INSTALL_DIR && $COMPOSE_CMD up -d"
    echo ""
fi

echo "Next steps:"
echo "  1. Add the postgresURL to your config file"
echo "  2. Run 'timesheetz' - it will auto-sync your data"
echo ""
