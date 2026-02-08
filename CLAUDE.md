# Timesheetz - Claude Instructions

## Overview

Timesheetz is a timesheet tracking TUI/web application for freelancers. It supports tracking client hours, vacation, training, sick time, and holidays, with features for client/rate management and earnings calculations.

## Architecture

- **TUI**: Bubble Tea-based terminal UI (`internal/ui/`)
- **API**: Gin REST server (`api/`)
- **Data Layer**: Pluggable backend supporting SQLite and PostgreSQL (`internal/db/`, `internal/datalayer/`)
- **Config**: YAML configuration with CLI flag and env var overrides (`internal/config/`)

## Features

- Daily timesheet entry (client hours, vacation, training, sick, holiday, idle)
- Client management with historical rate tracking
- Vacation carryover support
- Training budget tracking
- Earnings calculation with Euro formatting
- PDF/Excel export
- Multi-database support (SQLite local, PostgreSQL networked)

## Tech Stack

- Go 1.21+
- Bubble Tea (TUI)
- Gin (HTTP)
- SQLite (local) / PostgreSQL (network)
- GoReleaser + Homebrew

## Database Configuration

Supports two database backends:

### SQLite (Default)
```bash
timesheetz  # Uses ~/.local/share/timesheetz/timesheet.db
```

### PostgreSQL
```bash
# Via CLI flags
timesheetz --db-type postgres --postgres-url "postgres://user:pass@host:5432/db?sslmode=disable"

# Via environment variables
export TIMESHEETZ_DB_TYPE=postgres
export TIMESHEETZ_POSTGRES_URL="postgres://user:pass@host:5432/db?sslmode=disable"
timesheetz

# Via config file (~/.config/timesheetz/config.yaml)
dbType: postgres
postgresURL: "postgres://user:pass@host:5432/db?sslmode=disable"
```

Configuration precedence: CLI flags > env vars > config file > defaults

### PostgreSQL Setup
Use the setup scripts to create PostgreSQL on Docker or Kubernetes:
```bash
# Docker
./scripts/setup-postgres.sh --host <your-server-ip>

# Kubernetes  
./scripts/setup-postgres-k8s.sh --storage-class <your-storage-class>
```

## Release

```bash
# Tag and push triggers GoReleaser
git tag v1.x.x && git push --tags

# After release, update local
brew update && brew upgrade timesheetz
```

## Workflow

Before implementing any feature or change:
1. Rephrase the request to confirm understanding
2. Wait for user approval before proceeding
3. Only implement after explicit confirmation

## Testing

After implementing changes:
1. Check if tests need to be added, modified, or removed
2. Update tests to match the new behavior
3. Run `go test ./...` to verify all tests pass
4. Fix any failing tests before considering the work complete
