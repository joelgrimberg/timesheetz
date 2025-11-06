# Deployment Guide

This guide explains how to deploy the Timesheetz backend to a Docker server with Traefik.

## Prerequisites

- Docker and Docker Compose installed on the server
- Traefik already running and configured
- Access to the server via SSH or similar

## Deployment Steps

### 1. Copy Files to Server

Copy the following files to your server:
- `Dockerfile`
- `docker-compose.yml`
- `.dockerignore`
- The entire source code directory (or clone the repository)

### 2. Build and Start the Container

```bash
cd /path/to/timesheetz
docker-compose up -d --build
```

This will:
- Build the Docker image
- Start the container
- Mount the database volume
- Expose the API on port 8080

### 3. Configure Traefik

The `docker-compose.yml` includes Traefik labels. Make sure Traefik is configured to:
- Listen on the `web` entrypoint (or update the label to match your setup)
- Have access to the Docker socket for service discovery

The service will be available at `http://timesheetz.local` (or whatever hostname you configure).

### 4. Verify Deployment

Check that the API is running:
```bash
curl http://timesheetz.local/health
```

You should see:
```json
{"status":"ok"}
```

### 5. Database Migration

If you have an existing database on your MacBook:

1. Copy the database file from your MacBook:
   ```bash
   # On MacBook
   scp ~/.config/timesheetz/timesheet.db user@server:/tmp/timesheet.db
   ```

2. Copy it into the Docker volume:
   ```bash
   # On server
   docker cp /tmp/timesheet.db timesheetz-api:/app/data/timesheet.db
   ```

3. Restart the container:
   ```bash
   docker-compose restart
   ```

## Configuration

### Environment Variables

You can set the following environment variables in `docker-compose.yml`:

- `TIMESHEETZ_DB_PATH`: Path to the database file (default: `/app/data/timesheet.db`)

### Traefik Configuration

Update the Traefik labels in `docker-compose.yml` to match your setup:

- Change `Host(\`timesheetz.local\`)` to your desired hostname
- Update `entrypoints=web` if you use a different entrypoint
- Add TLS configuration if needed (uncomment the TLS labels)

## Maintenance

### View Logs

```bash
docker-compose logs -f timesheetz-api
```

### Stop the Service

```bash
docker-compose down
```

### Update the Service

```bash
docker-compose pull  # If using a registry
docker-compose up -d --build
```

### Backup Database

```bash
docker cp timesheetz-api:/app/data/timesheet.db ./backup-$(date +%Y%m%d).db
```

## Troubleshooting

### Container Won't Start

Check logs:
```bash
docker-compose logs timesheetz-api
```

### Database Permissions

Ensure the database file has correct permissions:
```bash
docker exec timesheetz-api chmod 644 /app/data/timesheet.db
```

### Traefik Not Routing

1. Check Traefik logs
2. Verify labels are correct
3. Ensure Traefik has access to Docker socket
4. Check that the service is healthy: `docker-compose ps`

