# Migration Guide: Local to Remote API

This guide explains how to migrate from local database access to a remote API using the dual-mode validation feature.

## Migration Strategy

The migration happens in three phases:

1. **Dual Mode**: Write to both local DB and remote API, compare reads
2. **Remote Mode**: Use only remote API
3. **Cleanup** (optional): Remove local database code

## Phase 1: Dual Mode (Validation)

### Step 1: Deploy Backend to Server

Follow the [Deployment Guide](./deployment.md) to deploy the API to your server.

### Step 2: Configure TUI for Dual Mode

Edit your `config.json` file (located at `~/.config/timesheetz/config.json`):

```json
{
  "apiMode": "dual",
  "apiBaseURL": "http://timesheetz.local"
}
```

Or set environment variables:
```bash
export TIMESHEETZ_API_MODE=dual
export TIMESHEETZ_API_URL=http://timesheetz.local
```

### Step 3: Run in Dual Mode

Start the TUI as normal:
```bash
timesheet
```

The application will now:
- Write all changes to both local database and remote API
- Read from both sources and compare results
- Log any discrepancies to help you identify issues

### Step 4: Monitor and Validate

1. Use the application normally for a period of time (days/weeks)
2. Check logs for any discrepancies:
   ```bash
   # Logs are written to the application log file
   tail -f ~/Library/Logs/timesheetz.out  # macOS
   ```

3. Verify data consistency:
   - Compare local and remote data manually if needed
   - Check that all writes succeed in both locations
   - Ensure reads match between local and remote

### Step 5: Resolve Any Issues

If you see discrepancies:
- Check network connectivity
- Verify API is accessible
- Review error logs
- Fix any data inconsistencies manually if needed

## Phase 2: Remote Mode

Once you're confident that dual mode is working correctly:

### Step 1: Switch to Remote Mode

Update `config.json`:
```json
{
  "apiMode": "remote",
  "apiBaseURL": "http://timesheetz.local"
}
```

Or set environment variable:
```bash
export TIMESHEETZ_API_MODE=remote
```

### Step 2: Test Remote Mode

Start the TUI:
```bash
timesheet
```

The application will now:
- Use only the remote API
- No longer access the local database
- All operations go through HTTP

### Step 3: Verify Everything Works

Test all functionality:
- Create new entries
- Update existing entries
- Delete entries
- View all views (timesheet, training, vacation, etc.)

### Step 4: Optional - Remove Local Database

Once you're confident remote mode works, you can optionally remove the local database:

```bash
# Backup first!
cp ~/.config/timesheetz/timesheet.db ~/.config/timesheetz/timesheet.db.backup

# Then remove
rm ~/.config/timesheetz/timesheet.db
```

## Rollback

If you need to rollback to local mode:

1. Update `config.json`:
   ```json
   {
     "apiMode": "local"
   }
   ```

2. Restart the application

The local database will be used again (if it still exists).

## Troubleshooting

### Dual Mode: Remote API Fails

If the remote API fails in dual mode:
- The application will continue using the local database
- Warnings will be logged
- Fix the API connection and try again

### Remote Mode: API Unavailable

If the API is unavailable in remote mode:
- The application will fail to start or operations will fail
- Check API connectivity
- Consider switching back to dual mode temporarily

### Data Mismatches in Dual Mode

If you see data mismatches:
1. Check which source has the correct data
2. Manually sync if needed
3. Investigate why the mismatch occurred
4. Fix the root cause before switching to remote mode

## Configuration Reference

### API Modes

- `"local"`: Use local database only (default, current behavior)
- `"dual"`: Write to both, read from both and compare
- `"remote"`: Use remote API only

### Configuration Options

In `config.json`:
- `apiMode`: One of "local", "dual", or "remote"
- `apiBaseURL`: Base URL for remote API (e.g., "http://timesheetz.local")

Environment variables (override config.json):
- `TIMESHEETZ_API_MODE`: API mode
- `TIMESHEETZ_API_URL`: API base URL

