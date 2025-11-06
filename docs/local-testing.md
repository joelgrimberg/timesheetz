# Local Testing with Docker

This guide explains how to test the remote API functionality locally using Docker, without deploying to a server.

## Overview

The setup allows you to:
1. Run the backend API and database in a Docker container
2. Run the TUI client on your local machine
3. Connect the local TUI to the containerized API in **dual mode** to validate data consistency

## Prerequisites

- Docker and Docker Compose installed
- Go installed (for building/running the local TUI)

## Quick Start

### 1. Start the Docker API Server

```bash
# Build and start the API server in Docker
docker-compose -f docker-compose.local.yml up -d

# Check if it's running
docker-compose -f docker-compose.local.yml ps

# View logs
docker-compose -f docker-compose.local.yml logs -f
```

The API will be available at `http://localhost:8080`

### 2. Test the API

```bash
# Check health endpoint
curl http://localhost:8080/health

# Should return: {"status":"ok"}
```

### 3. Configure Your Local TUI for Dual Mode

Update your local config file (`~/.config/timesheetz/config.json`) to enable dual mode:

```json
{
  "apiMode": "dual",
  "apiBaseURL": "http://localhost:8080",
  ...
}
```

Or set environment variables:

```bash
export TIMESHEETZ_API_MODE=dual
export TIMESHEETZ_API_URL=http://localhost:8080
```

### 4. Run Your Local TUI

```bash
# Run the TUI (it will connect to the Docker API in dual mode)
go run ./cmd/timesheet

# Or if you have it built:
./bin/timesheet
```

## How Dual Mode Works

In dual mode:
- **Writes** (Add, Update, Delete): Operations are performed on both your local database AND the remote API
- **Reads** (Get, GetAll): Operations are performed on both, and results are compared for validation
- **Validation**: Any discrepancies between local and remote data are logged

This allows you to:
- Verify that API calls work correctly
- Ensure data consistency between local and remote
- Catch any issues before fully migrating to remote mode

## Testing Workflow

1. **Start Docker API**: `docker-compose -f docker-compose.local.yml up -d`
2. **Set dual mode**: Configure your local app with `apiMode: "dual"` and `apiBaseURL: "http://localhost:8080"`
3. **Run local TUI**: Start your local app and perform operations
4. **Monitor logs**: Check both Docker logs and your local app logs for any discrepancies
5. **Validate data**: Compare data in both databases to ensure consistency

## Copying Your Local Database to Docker (Optional)

If you want to test with your existing data:

```bash
# Stop the container
docker-compose -f docker-compose.local.yml down

# Copy your local database
docker volume create timesheetz-db-local
docker run --rm -v ~/.local/share/timesheetz:/source -v timesheetz-db-local:/dest alpine sh -c "cp /source/timesheet.db /dest/timesheet.db"

# Start the container again
docker-compose -f docker-compose.local.yml up -d
```

## Stopping the Docker Container

```bash
# Stop and remove containers
docker-compose -f docker-compose.local.yml down

# Remove volumes (deletes database data)
docker-compose -f docker-compose.local.yml down -v
```

## Troubleshooting

### API not accessible

```bash
# Check if container is running
docker-compose -f docker-compose.local.yml ps

# Check logs
docker-compose -f docker-compose.local.yml logs

# Test health endpoint
curl http://localhost:8080/health
```

### Connection refused

- Make sure port 8080 is not already in use: `lsof -i :8080`
- Check Docker container logs: `docker-compose -f docker-compose.local.yml logs`

### Database issues

- The database is stored in a Docker volume: `timesheetz-db-local`
- To inspect: `docker volume inspect timesheetz-db-local`
- To reset: `docker-compose -f docker-compose.local.yml down -v`

## Next Steps

Once you've validated everything works in dual mode:
1. Switch to `"apiMode": "remote"` to use only the remote API
2. Deploy the Docker setup to your actual server
3. Update `apiBaseURL` to point to your server (e.g., `http://timesheetz.local`)

